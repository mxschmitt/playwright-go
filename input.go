package playwright

type mouseImpl struct {
	channel *channel
}

func newMouse(channel *channel) *mouseImpl {
	return &mouseImpl{
		channel: channel,
	}
}

func (m *mouseImpl) Move(x float64, y float64, options ...MouseMoveOptions) error {
	_, err := m.channel.Send("mouseMove", map[string]any{
		"x": x,
		"y": y,
	}, options)
	return err
}

func (m *mouseImpl) Down(options ...MouseDownOptions) error {
	_, err := m.channel.Send("mouseDown", options)
	return err
}

func (m *mouseImpl) Up(options ...MouseUpOptions) error {
	_, err := m.channel.Send("mouseUp", options)
	return err
}

func (m *mouseImpl) Click(x, y float64, options ...MouseClickOptions) error {
	_, err := m.channel.Send("mouseClick", map[string]any{
		"x": x,
		"y": y,
	}, options)
	return err
}

func (m *mouseImpl) Dblclick(x, y float64, options ...MouseDblclickOptions) error {
	var option MouseDblclickOptions
	if len(options) == 1 {
		option = options[0]
	}
	return m.Click(x, y, MouseClickOptions{
		ClickCount: Int(2),
		Button:     option.Button,
		Delay:      option.Delay,
	})
}

func (m *mouseImpl) Wheel(deltaX, deltaY float64) error {
	_, err := m.channel.Send("mouseWheel", map[string]any{
		"deltaX": deltaX,
		"deltaY": deltaY,
	})
	return err
}

type keyboardImpl struct {
	channel *channel
}

func newKeyboard(channel *channel) *keyboardImpl {
	return &keyboardImpl{
		channel: channel,
	}
}

func (m *keyboardImpl) Down(key string) error {
	_, err := m.channel.Send("keyboardDown", map[string]any{
		"key": key,
	})
	return err
}

func (m *keyboardImpl) Up(key string) error {
	_, err := m.channel.Send("keyboardUp", map[string]any{
		"key": key,
	})
	return err
}

func (m *keyboardImpl) InsertText(text string) error {
	_, err := m.channel.Send("keyboardInsertText", map[string]any{
		"text": text,
	})
	return err
}

func (m *keyboardImpl) Type(text string, options ...KeyboardTypeOptions) error {
	_, err := m.channel.Send("keyboardInsertText", map[string]any{
		"text": text,
	}, options)
	return err
}

func (m *keyboardImpl) Press(key string, options ...KeyboardPressOptions) error {
	_, err := m.channel.Send("keyboardPress", map[string]any{
		"key": key,
	}, options)
	return err
}

type touchscreenImpl struct {
	channel *channel
}

func newTouchscreen(channel *channel) *touchscreenImpl {
	return &touchscreenImpl{
		channel: channel,
	}
}

func (t *touchscreenImpl) Tap(x int, y int) error {
	_, err := t.channel.Send("touchscreenTap", map[string]any{"x": x, "y": y})
	return err
}
