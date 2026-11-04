package playwright

type webMCPImpl struct {
	frame *frameImpl
}

func newWebMCP(frame *frameImpl) *webMCPImpl {
	return &webMCPImpl{frame: frame}
}

func (w *webMCPImpl) Tools(options ...WebMCPToolsOptions) ([]WebMCPTool, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	result, err := w.frame.channel.SendWithTimeout("webmcpTools", resolveTimeout(w.frame.page.timeoutSettings, explicit), options)
	if err != nil {
		return nil, err
	}
	items, _ := result.([]any)
	tools := make([]WebMCPTool, len(items))
	for i, item := range items {
		remapMapToStruct(item, &tools[i])
	}
	return tools, nil
}

func (w *webMCPImpl) CallTool(name string, input any, options ...WebMCPCallToolOptions) (any, error) {
	var explicit *float64
	if len(options) == 1 {
		explicit = options[0].Timeout
	}
	result, err := w.frame.channel.SendWithTimeout("webmcpCallTool", resolveTimeout(w.frame.page.timeoutSettings, explicit), map[string]any{
		"name":  name,
		"input": serializeArgument(input),
	}, options)
	if err != nil {
		return nil, err
	}
	return parseResult(result), nil
}
