package playwright

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type routeImpl struct {
	channelOwner
	handling *chan bool
	context  *browserContextImpl
	didThrow bool
}

func (r *routeImpl) startHandling() chan bool {
	r.Lock()
	defer r.Unlock()
	handling := make(chan bool, 1)
	r.handling = &handling
	return handling
}

func (r *routeImpl) reportHandled(done bool) {
	r.Lock()
	defer r.Unlock()
	if r.handling != nil {
		handling := *r.handling
		r.handling = nil
		handling <- done
	}
}

func (r *routeImpl) checkNotHandled() error {
	r.RLock()
	defer r.RUnlock()
	if r.handling == nil {
		return errors.New("Route is already handled!")
	}
	return nil
}

func (r *routeImpl) Request() Request {
	return fromChannel(r.initializer["request"]).(*requestImpl)
}

func unpackOptionalArgument(input any) any {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Slice {
		panic("Needs to be a slice")
	}
	if inputValue.Len() == 0 {
		return nil
	}
	return inputValue.Index(0).Interface()
}

func (r *routeImpl) Abort(errorCode ...string) error {
	return r.handleRoute(func() error {
		return r.raceWithPageClose(func() error {
			_, err := r.channel.Send("abort", map[string]any{
				"errorCode": unpackOptionalArgument(errorCode),
			})
			return err
		})
	})
}

func (r *routeImpl) raceWithPageClose(f func() error) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- f()
	}()

	select {
	case <-r.Request().(*requestImpl).targetClosed():
		// upstream does not throw the err
		return nil
	case err := <-errChan:
		return err
	}
}

func (r *routeImpl) Fulfill(options ...RouteFulfillOptions) error {
	return r.handleRoute(func() error {
		return r.innerFulfill(options...)
	})
}

func (r *routeImpl) handleRoute(cb func() error) error {
	err := r.checkNotHandled()
	if err != nil {
		return err
	}
	if err := cb(); err != nil {
		r.didThrow = true
		return err
	}
	r.reportHandled(true)
	return nil
}

func (r *routeImpl) innerFulfill(options ...RouteFulfillOptions) error {
	err := r.checkNotHandled()
	if err != nil {
		return err
	}
	option := RouteFulfillOptions{}
	if len(options) == 1 {
		option = options[0]
	}
	overrides := map[string]any{
		"status": 200,
	}
	headers := make(map[string]string)

	if option.Response != nil {
		overrides["status"] = option.Response.Status()
		headers = option.Response.Headers()
		response, ok := option.Response.(*apiResponseImpl)
		if option.Body == nil && option.Path == nil && ok && response.request.connection == r.connection {
			overrides["fetchResponseUid"] = response.fetchUid()
		} else {
			option.Body, _ = option.Response.Body()
		}
		option.Response = nil
	}
	if option.Status != nil {
		overrides["status"] = *option.Status
		option.Status = nil
	}

	length := 0
	isBase64 := false
	var fileContentType string
	if _, ok := option.Body.(string); ok {
		isBase64 = false
	} else if body, ok := option.Body.([]byte); ok {
		option.Body = base64.StdEncoding.EncodeToString(body)
		length = len(body)
		isBase64 = true
	} else if option.Path != nil {
		content, err := os.ReadFile(*option.Path)
		if err != nil {
			return err
		}
		fileContentType = http.DetectContentType(content)
		option.Body = base64.StdEncoding.EncodeToString(content)
		isBase64 = true
		length = len(content)
	}

	if option.Headers != nil {
		headers = make(map[string]string)
		for key, val := range option.Headers {
			headers[strings.ToLower(key)] = val
		}
		option.Headers = nil
	}
	if option.ContentType != nil {
		headers["content-type"] = *option.ContentType
		option.ContentType = nil
	} else if option.Path != nil {
		headers["content-type"] = fileContentType
	}
	if _, ok := headers["content-length"]; !ok && length > 0 {
		headers["content-length"] = strconv.Itoa(length)
	}
	overrides["headers"] = serializeMapToNameAndValue(headers)
	overrides["isBase64"] = isBase64

	option.Path = nil
	err = r.raceWithPageClose(func() error {
		_, err := r.channel.Send("fulfill", option, overrides)
		return err
	})
	return err
}

func (r *routeImpl) Fallback(options ...RouteFallbackOptions) error {
	err := r.checkNotHandled()
	if err != nil {
		return err
	}
	opt := RouteFallbackOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	r.Request().(*requestImpl).applyFallbackOverrides(opt)
	r.reportHandled(false)
	return nil
}

func (r *routeImpl) Fetch(options ...RouteFetchOptions) (APIResponse, error) {
	opt := &APIRequestContextFetchOptions{}
	url := ""
	if len(options) == 1 {
		err := assignStructFields(opt, options[0], true)
		if err != nil {
			return nil, err
		}
		opt.Data = options[0].PostData
		if options[0].URL != nil {
			url = *options[0].URL
		}
	}
	ret, err := r.connection.WrapAPICall(func() (any, error) {
		return r.context.request.innerFetch(url, r.Request(), *opt)
	}, false)
	if ret == nil {
		return nil, err
	}
	return ret.(APIResponse), err
}

func (r *routeImpl) Continue(options ...RouteContinueOptions) error {
	option := &RouteFallbackOptions{}
	if len(options) == 1 {
		err := assignStructFields(option, options[0], true)
		if err != nil {
			return err
		}
	}

	return r.handleRoute(func() error {
		r.Request().(*requestImpl).applyFallbackOverrides(*option)
		return r.internalContinue(false)
	})
}

func (r *routeImpl) internalContinue(isFallback bool) error {
	overrides := make(map[string]any)
	overrides["url"] = r.Request().(*requestImpl).fallbackOverrides.URL
	overrides["method"] = r.Request().(*requestImpl).fallbackOverrides.Method
	headers := r.Request().(*requestImpl).fallbackOverrides.Headers
	if headers != nil {
		overrides["headers"] = serializeMapToNameAndValue(headers)
	}
	postDataBuf := r.Request().(*requestImpl).fallbackOverrides.PostDataBuffer
	if postDataBuf != nil {
		overrides["postData"] = base64.StdEncoding.EncodeToString(postDataBuf)
	}
	overrides["isFallback"] = isFallback
	return r.raceWithPageClose(func() error {
		_, err := r.channel.Send("continue", overrides)
		return err
	})
}

func (r *routeImpl) redirectedNavigationRequest(url string) error {
	return r.handleRoute(func() error {
		return r.raceWithPageClose(func() error {
			_, err := r.channel.Send("redirectNavigationRequest", map[string]any{
				"url": url,
			})
			return err
		})
	})
}

func newRoute(parent *channelOwner, objectType string, guid string, initializer map[string]any) *routeImpl {
	bt := &routeImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	bt.markAsInternalType()
	return bt
}
