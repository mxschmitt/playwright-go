package playwright

import "encoding/base64"

type screencastImpl struct {
	page *pageImpl
}

func (s *screencastImpl) Start(options ...ScreencastStartOptions) error {
	overrides := map[string]any{}
	if len(options) == 1 {
		if options[0].OnFrame != nil {
			onFrame := options[0].OnFrame
			s.page.channel.On("screencastFrame", func(params map[string]any) {
				data, _ := base64.StdEncoding.DecodeString(params["data"].(string))
				onFrame(OnFrame{Data: data})
			})
			overrides["sendFrames"] = true
			options[0].OnFrame = nil // don't serialize the callback
		}
		if options[0].Path != nil {
			overrides["record"] = true
		}
	}
	_, err := s.page.channel.Send("screencastStart", options, overrides)
	return err
}

func (s *screencastImpl) Stop() error {
	_, err := s.page.channel.Send("screencastStop")
	return err
}

func (s *screencastImpl) ShowActions(options ...ScreencastShowActionsOptions) error {
	_, err := s.page.channel.Send("screencastShowActions", options)
	return err
}

func (s *screencastImpl) HideActions() error {
	_, err := s.page.channel.Send("screencastHideActions")
	return err
}

func (s *screencastImpl) ShowOverlay(html string, options ...ScreencastShowOverlayOptions) error {
	overrides := map[string]any{"html": html}
	_, err := s.page.channel.Send("screencastShowOverlay", options, overrides)
	return err
}

func (s *screencastImpl) ShowChapter(title string, options ...ScreencastShowChapterOptions) error {
	overrides := map[string]any{"title": title}
	_, err := s.page.channel.Send("screencastChapter", options, overrides)
	return err
}

func (s *screencastImpl) ShowOverlays() error {
	_, err := s.page.channel.Send("screencastSetOverlayVisible", map[string]any{"visible": true})
	return err
}

func (s *screencastImpl) HideOverlays() error {
	_, err := s.page.channel.Send("screencastSetOverlayVisible", map[string]any{"visible": false})
	return err
}
