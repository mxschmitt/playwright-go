package playwright

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type routeImpl struct {
	channelOwner
}

func (r *routeImpl) Request() Request {
	return fromChannel(r.initializer["request"]).(*requestImpl)
}

type RouteAbortOptions struct {
	ErrorCode *string `json:"errorCode"`
}

func (r *routeImpl) Abort(options ...RouteAbortOptions) error {
	_, err := r.channel.Send("abort", options)
	return err
}

type RouteFulfillOptions struct {
	Status      *int              `json:"status"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	Path        *string           `json:"path"`
	ContentType *string           `json:"contentType"`
}

func (r *routeImpl) Fulfill(options RouteFulfillOptions) error {
	length := 0
	isBase64 := false
	var fileContentType string
	if _, ok := options.Body.(string); ok {
		isBase64 = false
	} else if body, ok := options.Body.([]byte); ok {
		options.Body = base64.StdEncoding.EncodeToString(body)
		length = len(body)
		isBase64 = true
	} else if options.Path != nil {
		content, err := ioutil.ReadFile(*options.Path)
		if err != nil {
			return err
		}
		fileContentType = http.DetectContentType(content)
		options.Body = base64.StdEncoding.EncodeToString(content)
		isBase64 = true
		length = len(content)
	}

	headers := make(map[string]string)
	if options.Headers != nil {
		for key, val := range options.Headers {
			headers[strings.ToLower(key)] = val
		}
		options.Headers = nil
	}
	if options.ContentType != nil {
		headers["content-type"] = *options.ContentType
	} else if options.Path != nil {
		headers["content-type"] = fileContentType
	}
	if _, ok := headers["content-length"]; !ok && length > 0 {
		headers["content-length"] = strconv.Itoa(length)
	}

	options.Path = nil
	_, err := r.channel.Send("fulfill", options, map[string]interface{}{
		"isBase64": isBase64,
		"headers":  serializeHeaders(headers),
	})
	return err
}

func (r *routeImpl) Continue(options ...RouteContinueOptions) error {
	overrides := make(map[string]interface{})
	if len(options) == 1 {
		option := options[0]
		if option.Url != nil {
			overrides["url"] = option.Url
		}
		if option.Method != nil {
			overrides["method"] = option.Method
		}
		if option.Headers != nil {
			overrides["headers"] = serializeHeaders(option.Headers)
		}
		if option.PostData != nil {
			switch v := option.PostData.(type) {
			case string:
				overrides["postData"] = base64.StdEncoding.EncodeToString([]byte(v))
			case []byte:
				overrides["postData"] = base64.StdEncoding.EncodeToString(v)
			}
		}
	}
	_, err := r.channel.Send("continue", overrides)
	return err
}

func newRoute(parent *channelOwner, objectType string, guid string, initializer map[string]interface{}) *routeImpl {
	bt := &routeImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	return bt
}
