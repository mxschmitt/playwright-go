package playwright

import (
	"encoding/base64"
	"errors"
)

type screencastImpl struct {
	page     *pageImpl
	started  bool
	savePath *string
	artifact *artifactImpl
}

func (s *screencastImpl) Start(options ...ScreencastStartOptions) error {
	if s.started {
		return errors.New("Screencast is already started")
	}
	s.started = true
	overrides := map[string]any{}
	if len(options) == 1 {
		if options[0].OnFrame != nil {
			onFrame := options[0].OnFrame
			s.page.channel.On("screencastFrame", func(params map[string]any) {
				data, _ := base64.StdEncoding.DecodeString(params["data"].(string))
				frame := OnFrame{Data: data}
				if vw, ok := params["viewportWidth"].(float64); ok {
					frame.ViewportWidth = int(vw)
				}
				if vh, ok := params["viewportHeight"].(float64); ok {
					frame.ViewportHeight = int(vh)
				}
				onFrame(frame)
			})
			overrides["sendFrames"] = true
			options[0].OnFrame = nil // don't serialize the callback
		}
		if options[0].Path != nil {
			overrides["record"] = true
			s.savePath = options[0].Path
		}
	}
	result, err := s.page.channel.Send("screencastStart", options, overrides)
	if err != nil {
		return err
	}
	if resultMap, ok := result.(map[string]any); ok {
		if artifactChannel := fromNullableChannel(resultMap["artifact"]); artifactChannel != nil {
			s.artifact = artifactChannel.(*artifactImpl)
		}
	}
	return nil
}

func (s *screencastImpl) Stop() error {
	s.started = false
	if _, err := s.page.channel.Send("screencastStop"); err != nil {
		return err
	}
	if s.savePath != nil && s.artifact != nil {
		if err := s.artifact.SaveAs(*s.savePath); err != nil {
			return err
		}
	}
	s.artifact = nil
	s.savePath = nil
	return nil
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
