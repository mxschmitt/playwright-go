package playwright

import "sync"

type debuggerImpl struct {
	channelOwner
	mu            sync.RWMutex
	pausedDetails *PausedDetail
}

func (d *debuggerImpl) OnPausedStateChanged(fn func()) {
	d.On("pausedStateChanged", fn)
}

func (d *debuggerImpl) PausedDetails() (*PausedDetail, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pausedDetails, nil
}

func (d *debuggerImpl) RequestPause() error {
	_, err := d.channel.Send("requestPause")
	return err
}

func (d *debuggerImpl) Resume() error {
	_, err := d.channel.Send("resume")
	return err
}

func (d *debuggerImpl) Next() error {
	_, err := d.channel.Send("next")
	return err
}

func (d *debuggerImpl) RunTo(location DebuggerLocation) error {
	_, err := d.channel.Send("runTo", map[string]any{"location": location})
	return err
}

func newDebugger(parent *channelOwner, objectType string, guid string, initializer map[string]any) *debuggerImpl {
	bt := &debuggerImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	bt.channel.On("pausedStateChanged", func(params map[string]any) {
		bt.mu.Lock()
		if pd, ok := params["pausedDetails"]; ok && pd != nil {
			data := pd.(map[string]any)
			detail := &PausedDetail{Title: data["title"].(string)}
			if loc, ok := data["location"].(map[string]any); ok {
				detail.Location = &PausedDetailLocation{File: loc["file"].(string)}
				if line, ok := loc["line"]; ok && line != nil {
					v := int(line.(float64))
					detail.Location.Line = &v
				}
				if col, ok := loc["column"]; ok && col != nil {
					v := int(col.(float64))
					detail.Location.Column = &v
				}
			}
			bt.pausedDetails = detail
		} else {
			bt.pausedDetails = nil
		}
		bt.mu.Unlock()
		bt.Emit("pausedStateChanged")
	})
	return bt
}
