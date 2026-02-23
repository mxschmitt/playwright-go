package playwright

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type localUtilsImpl struct {
	channelOwner
	Devices map[string]*DeviceDescriptor
}

type (
	localUtilsZipOptions struct {
		ZipFile        string `json:"zipFile"`
		Entries        []any  `json:"entries"`
		StacksId       string `json:"stacksId"`
		Mode           string `json:"mode"`
		IncludeSources bool   `json:"includeSources"`
	}

	harLookupOptions struct {
		HarId               string            `json:"harId"`
		URL                 string            `json:"url"`
		Method              string            `json:"method"`
		Headers             map[string]string `json:"headers"`
		IsNavigationRequest bool              `json:"isNavigationRequest"`
		PostData            any               `json:"postData,omitempty"`
	}

	harLookupResult struct {
		Action      string              `json:"action"`
		Message     *string             `json:"message,omitempty"`
		RedirectURL *string             `json:"redirectUrl,omitempty"`
		Status      *int                `json:"status,omitempty"`
		Headers     []map[string]string `json:"headers,omitempty"`
		Body        *string             `json:"body,omitempty"`
	}
)

func (l *localUtilsImpl) Zip(options localUtilsZipOptions) (any, error) {
	return l.channel.Send("zip", options)
}

func (l *localUtilsImpl) HarOpen(file string) (string, error) {
	result, err := l.channel.SendReturnAsDict("harOpen", []map[string]any{
		{
			"file": file,
		},
	})
	if err == nil {
		if harId, ok := result["harId"]; ok {
			return harId.(string), nil
		}
		if err, ok := result["error"]; ok {
			return "", fmt.Errorf("%w:%v", ErrPlaywright, err)
		}
	}
	return "", err
}

func (l *localUtilsImpl) HarLookup(option harLookupOptions) (*harLookupResult, error) {
	overrides := make(map[string]any)
	overrides["harId"] = option.HarId
	overrides["url"] = option.URL
	overrides["method"] = option.Method
	if option.Headers != nil {
		overrides["headers"] = serializeMapToNameAndValue(option.Headers)
	}
	overrides["isNavigationRequest"] = option.IsNavigationRequest
	if option.PostData != nil {
		switch v := option.PostData.(type) {
		case string:
			overrides["postData"] = base64.StdEncoding.EncodeToString([]byte(v))
		case []byte:
			overrides["postData"] = base64.StdEncoding.EncodeToString(v)
		}
	}
	ret, err := l.channel.SendReturnAsDict("harLookup", overrides)
	if ret == nil {
		return nil, err
	}
	var result harLookupResult
	mJson, err := json.Marshal(ret)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(mJson, &result)
	if err != nil {
		return nil, err
	}
	if result.Body != nil {
		body, err := base64.StdEncoding.DecodeString(*result.Body)
		if err != nil {
			return nil, err
		}
		result.Body = String(string(body))
	}
	return &result, err
}

func (l *localUtilsImpl) HarClose(harId string) error {
	_, err := l.channel.Send("harClose", []map[string]any{
		{
			"harId": harId,
		},
	})
	return err
}

func (l *localUtilsImpl) HarUnzip(zipFile, harFile string) error {
	_, err := l.channel.Send("harUnzip", []map[string]any{
		{
			"zipFile": zipFile,
			"harFile": harFile,
		},
	})
	return err
}

func (l *localUtilsImpl) TracingStarted(traceName string, tracesDir ...string) (string, error) {
	overrides := make(map[string]any)
	overrides["traceName"] = traceName
	if len(tracesDir) > 0 {
		overrides["tracesDir"] = tracesDir[0]
	}
	stacksId, err := l.channel.Send("tracingStarted", overrides)
	if stacksId == nil {
		return "", err
	}
	return stacksId.(string), err
}

func (l *localUtilsImpl) TraceDiscarded(stacksId string) error {
	_, err := l.channel.Send("traceDiscarded", map[string]any{
		"stacksId": stacksId,
	})
	return err
}

func (l *localUtilsImpl) AddStackToTracingNoReply(id uint32, stack []map[string]any) {
	l.channel.SendNoReply("addStackToTracingNoReply", map[string]any{
		"callData": map[string]any{
			"id":    id,
			"stack": stack,
		},
	})
}

func newLocalUtils(parent *channelOwner, objectType string, guid string, initializer map[string]any) *localUtilsImpl {
	l := &localUtilsImpl{
		Devices: make(map[string]*DeviceDescriptor),
	}
	l.createChannelOwner(l, parent, objectType, guid, initializer)
	for _, dd := range initializer["deviceDescriptors"].([]any) {
		entry := dd.(map[string]any)
		l.Devices[entry["name"].(string)] = &DeviceDescriptor{
			Viewport: &Size{},
		}
		remapMapToStruct(entry["descriptor"], l.Devices[entry["name"].(string)])
	}
	l.markAsInternalType()
	return l
}
