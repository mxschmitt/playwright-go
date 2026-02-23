package playwright

type dialogImpl struct {
	channelOwner
	page Page
}

func (d *dialogImpl) Type() string {
	return d.initializer["type"].(string)
}

func (d *dialogImpl) Message() string {
	return d.initializer["message"].(string)
}

func (d *dialogImpl) DefaultValue() string {
	return d.initializer["defaultValue"].(string)
}

func (d *dialogImpl) Accept(promptTextInput ...string) error {
	var promptText *string
	if len(promptTextInput) == 1 {
		promptText = &promptTextInput[0]
	}
	_, err := d.channel.Send("accept", map[string]any{
		"promptText": promptText,
	})
	return err
}

func (d *dialogImpl) Dismiss() error {
	_, err := d.channel.Send("dismiss")
	return err
}

func (d *dialogImpl) Page() Page {
	return d.page
}

func newDialog(parent *channelOwner, objectType string, guid string, initializer map[string]any) *dialogImpl {
	bt := &dialogImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	page := fromNullableChannel(initializer["page"])
	if page != nil {
		bt.page = page.(*pageImpl)
	}
	return bt
}
