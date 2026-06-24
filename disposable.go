package playwright

type disposableImpl struct {
	channelOwner
}

func (d *disposableImpl) Dispose() error {
	_, err := d.channel.Send("dispose")
	return err
}

func newDisposable(parent *channelOwner, objectType string, guid string, initializer map[string]any) *disposableImpl {
	bt := &disposableImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	return bt
}
