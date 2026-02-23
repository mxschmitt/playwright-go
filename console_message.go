package playwright

type consoleMessageImpl struct {
	event  map[string]any
	page   Page
	worker Worker
}

func (c *consoleMessageImpl) Type() string {
	return c.event["type"].(string)
}

func (c *consoleMessageImpl) Text() string {
	return c.event["text"].(string)
}

func (c *consoleMessageImpl) String() string {
	return c.Text()
}

func (c *consoleMessageImpl) Args() []JSHandle {
	args := c.event["args"].([]any)
	out := []JSHandle{}
	for idx := range args {
		out = append(out, fromChannel(args[idx]).(*jsHandleImpl))
	}
	return out
}

func (c *consoleMessageImpl) Location() *ConsoleMessageLocation {
	location := &ConsoleMessageLocation{}
	remapMapToStruct(c.event["location"], location)
	return location
}

func (c *consoleMessageImpl) Page() Page {
	return c.page
}

func (c *consoleMessageImpl) Worker() (Worker, error) {
	return c.worker, nil
}

func newConsoleMessage(event map[string]any) *consoleMessageImpl {
	bt := &consoleMessageImpl{}
	bt.event = event
	page := fromNullableChannel(event["page"])
	if page != nil {
		bt.page = page.(*pageImpl)
	}
	worker := fromNullableChannel(event["worker"])
	if worker != nil {
		bt.worker = worker.(*workerImpl)
	}
	return bt
}
