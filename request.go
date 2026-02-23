package playwright

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type serializedFallbackOverrides struct {
	URL            *string
	Method         *string
	Headers        map[string]string
	PostDataBuffer []byte
}

type requestImpl struct {
	channelOwner
	timing             *RequestTiming
	provisionalHeaders *rawHeaders
	allHeaders         *rawHeaders
	redirectedFrom     Request
	redirectedTo       Request
	failureText        string
	fallbackOverrides  *serializedFallbackOverrides
}

func (r *requestImpl) URL() string {
	if r.fallbackOverrides.URL != nil {
		return *r.fallbackOverrides.URL
	}
	return r.initializer["url"].(string)
}

func (r *requestImpl) ResourceType() string {
	return r.initializer["resourceType"].(string)
}

func (r *requestImpl) Method() string {
	if r.fallbackOverrides.Method != nil {
		return *r.fallbackOverrides.Method
	}
	return r.initializer["method"].(string)
}

func (r *requestImpl) PostData() (string, error) {
	body, err := r.PostDataBuffer()
	if err != nil {
		return "", err
	}
	return string(body), err
}

func (r *requestImpl) PostDataJSON(v any) error {
	body, err := r.PostDataBuffer()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (r *requestImpl) PostDataBuffer() ([]byte, error) {
	if r.fallbackOverrides.PostDataBuffer != nil {
		return r.fallbackOverrides.PostDataBuffer, nil
	}
	if _, ok := r.initializer["postData"]; !ok {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(r.initializer["postData"].(string))
}

func (r *requestImpl) Headers() map[string]string {
	if r.fallbackOverrides.Headers != nil {
		return newRawHeaders(serializeMapToNameAndValue(r.fallbackOverrides.Headers)).Headers()
	}
	return r.provisionalHeaders.Headers()
}

func (r *requestImpl) Response() (Response, error) {
	channel, err := r.channel.Send("response")
	if err != nil {
		return nil, err
	}
	channelOwner := fromNullableChannel(channel)
	if channelOwner == nil {
		// no response
		return nil, nil
	}
	return channelOwner.(*responseImpl), nil
}

func (r *requestImpl) Frame() Frame {
	channel, ok := r.initializer["frame"]
	if !ok {
		// Service Worker requests do not have an associated frame.
		return nil
	}
	frame := fromChannel(channel).(*frameImpl)
	if frame.page == nil {
		// Frame for this navigation request is not available, because the request
		// was issued before the frame is created. You can check whether the request
		// is a navigation request by calling IsNavigationRequest() method.
		return nil
	}
	return frame
}

func (r *requestImpl) IsNavigationRequest() bool {
	return r.initializer["isNavigationRequest"].(bool)
}

func (r *requestImpl) RedirectedFrom() Request {
	return r.redirectedFrom
}

func (r *requestImpl) RedirectedTo() Request {
	return r.redirectedTo
}

func (r *requestImpl) Failure() error {
	if r.failureText == "" {
		return nil
	}
	return fmt.Errorf("%v", r.failureText)
}

func (r *requestImpl) Timing() *RequestTiming {
	return r.timing
}

func (r *requestImpl) AllHeaders() (map[string]string, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return nil, err
	}
	return headers.Headers(), nil
}

func (r *requestImpl) HeadersArray() ([]NameValue, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return nil, err
	}
	return headers.HeadersArray(), nil
}

func (r *requestImpl) HeaderValue(name string) (string, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return "", err
	}
	return headers.Get(name), err
}

func (r *requestImpl) HeaderValues(name string) ([]string, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return []string{}, err
	}
	return headers.GetAll(name), err
}

func (r *requestImpl) ActualHeaders() (*rawHeaders, error) {
	if r.fallbackOverrides.Headers != nil {
		return newRawHeaders(serializeMapToNameAndValue(r.fallbackOverrides.Headers)), nil
	}
	if r.allHeaders == nil {
		response, err := r.Response()
		if err != nil {
			return nil, err
		}
		if response == nil {
			return r.provisionalHeaders, nil
		}
		headers, err := r.channel.Send("rawRequestHeaders")
		if err != nil {
			return nil, err
		}
		r.allHeaders = newRawHeaders(headers)
	}
	return r.allHeaders, nil
}

func (r *requestImpl) ServiceWorker() Worker {
	channel, ok := r.initializer["serviceWorker"]
	if !ok {
		return nil
	}
	return fromChannel(channel).(*workerImpl)
}

func (r *requestImpl) Sizes() (*RequestSizesResult, error) {
	response, err := r.Response()
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("sizes could not be retrieved because request has no response")
	}
	sizes, err := response.(*responseImpl).channel.Send("sizes")
	if err != nil {
		return nil, err
	}
	result := &RequestSizesResult{}
	remapMapToStruct(sizes, result)
	return result, nil
}

func (r *requestImpl) applyFallbackOverrides(options RouteFallbackOptions) {
	if options.URL != nil {
		r.fallbackOverrides.URL = options.URL
	}
	if options.Method != nil {
		r.fallbackOverrides.Method = options.Method
	}
	r.fallbackOverrides.Headers = options.Headers
	if options.PostData != nil {
		switch v := options.PostData.(type) {
		case string:
			r.fallbackOverrides.PostDataBuffer = []byte(v)
		case []byte:
			r.fallbackOverrides.PostDataBuffer = v
		}
	}
}

func (r *requestImpl) targetClosed() <-chan error {
	page := r.safePage()
	if page == nil {
		return make(<-chan error, 1)
	}
	return page.closedOrCrashed
}

func (r *requestImpl) setResponseEndTiming(t float64) {
	r.timing.ResponseEnd = t
	if r.timing.ResponseStart == -1 {
		r.timing.ResponseStart = t
	}
}

func (r *requestImpl) safePage() *pageImpl {
	channel := fromNullableChannel(r.initializer["frame"])
	if channel == nil {
		return nil
	}
	frame, ok := channel.(*frameImpl)
	if !ok {
		return nil
	}
	return frame.page
}

func newRequest(parent *channelOwner, objectType string, guid string, initializer map[string]any) *requestImpl {
	req := &requestImpl{}
	req.createChannelOwner(req, parent, objectType, guid, initializer)
	redirectedFrom := fromNullableChannel(initializer["redirectedFrom"])
	if redirectedFrom != nil {
		req.redirectedFrom = redirectedFrom.(*requestImpl)
	}
	if req.redirectedFrom != nil {
		req.redirectedFrom.(*requestImpl).redirectedTo = req
	}
	req.timing = &RequestTiming{
		StartTime:             0,
		DomainLookupStart:     -1,
		DomainLookupEnd:       -1,
		ConnectStart:          -1,
		SecureConnectionStart: -1,
		ConnectEnd:            -1,
		RequestStart:          -1,
		ResponseStart:         -1,
		ResponseEnd:           -1,
	}
	req.provisionalHeaders = newRawHeaders(req.initializer["headers"])
	req.fallbackOverrides = &serializedFallbackOverrides{}
	return req
}
