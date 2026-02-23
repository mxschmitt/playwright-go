package playwright

import (
	"encoding/base64"
	"encoding/json"
)

type responseImpl struct {
	channelOwner
	request            *requestImpl
	provisionalHeaders *rawHeaders
	rawHeaders         *rawHeaders
	finished           chan error
}

func (r *responseImpl) FromServiceWorker() bool {
	return r.initializer["fromServiceWorker"].(bool)
}

func (r *responseImpl) URL() string {
	return r.initializer["url"].(string)
}

func (r *responseImpl) Ok() bool {
	return r.Status() == 0 || (r.Status() >= 200 && r.Status() <= 299)
}

func (r *responseImpl) Status() int {
	return int(r.initializer["status"].(float64))
}

func (r *responseImpl) StatusText() string {
	return r.initializer["statusText"].(string)
}

func (r *responseImpl) Headers() map[string]string {
	return r.provisionalHeaders.Headers()
}

func (r *responseImpl) Finished() error {
	select {
	case err := <-r.request.targetClosed():
		return err
	case err := <-r.finished:
		return err
	}
}

func (r *responseImpl) Body() ([]byte, error) {
	b64Body, err := r.channel.Send("body")
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(b64Body.(string))
}

func (r *responseImpl) Text() (string, error) {
	body, err := r.Body()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (r *responseImpl) JSON(v any) error {
	body, err := r.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (r *responseImpl) Request() Request {
	return r.request
}

func (r *responseImpl) Frame() Frame {
	return r.request.Frame()
}

func (r *responseImpl) AllHeaders() (map[string]string, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return nil, err
	}
	return headers.Headers(), nil
}

func (r *responseImpl) HeadersArray() ([]NameValue, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return nil, err
	}
	return headers.HeadersArray(), nil
}

func (r *responseImpl) HeaderValue(name string) (string, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return "", err
	}
	return headers.Get(name), err
}

func (r *responseImpl) HeaderValues(name string) ([]string, error) {
	headers, err := r.ActualHeaders()
	if err != nil {
		return []string{}, err
	}
	return headers.GetAll(name), err
}

func (r *responseImpl) ActualHeaders() (*rawHeaders, error) {
	if r.rawHeaders == nil {
		headers, err := r.channel.Send("rawResponseHeaders")
		if err != nil {
			return nil, err
		}
		r.rawHeaders = newRawHeaders(headers)
	}
	return r.rawHeaders, nil
}

func (r *responseImpl) SecurityDetails() (*ResponseSecurityDetailsResult, error) {
	details, err := r.channel.Send("securityDetails")
	if err != nil {
		return nil, err
	}
	result := &ResponseSecurityDetailsResult{}
	remapMapToStruct(details.(map[string]any), result)
	return result, nil
}

func (r *responseImpl) ServerAddr() (*ResponseServerAddrResult, error) {
	addr, err := r.channel.Send("serverAddr")
	if err != nil {
		return nil, err
	}
	result := &ResponseServerAddrResult{}
	remapMapToStruct(addr, result)
	return result, nil
}

func newResponse(parent *channelOwner, objectType string, guid string, initializer map[string]any) *responseImpl {
	resp := &responseImpl{}
	resp.createChannelOwner(resp, parent, objectType, guid, initializer)
	timing := resp.initializer["timing"].(map[string]any)
	resp.request = fromChannel(resp.initializer["request"]).(*requestImpl)
	resp.request.timing = &RequestTiming{
		StartTime:             timing["startTime"].(float64),
		DomainLookupStart:     timing["domainLookupStart"].(float64),
		DomainLookupEnd:       timing["domainLookupEnd"].(float64),
		ConnectStart:          timing["connectStart"].(float64),
		SecureConnectionStart: timing["secureConnectionStart"].(float64),
		ConnectEnd:            timing["connectEnd"].(float64),
		RequestStart:          timing["requestStart"].(float64),
		ResponseStart:         timing["responseStart"].(float64),
	}
	resp.provisionalHeaders = newRawHeaders(resp.initializer["headers"])
	resp.finished = make(chan error, 1)
	return resp
}
