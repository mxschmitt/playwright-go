package playwright

import (
	"errors"
	"time"
)

type clockImpl struct {
	browserCtx *browserContextImpl
}

func newClock(bCtx *browserContextImpl) Clock {
	return &clockImpl{
		browserCtx: bCtx,
	}
}

func (c *clockImpl) FastForward(ticks any) error {
	params, err := parseTicks(ticks)
	if err != nil {
		return err
	}

	_, err = c.browserCtx.channel.Send("clockFastForward", params)
	return err
}

func (c *clockImpl) Install(options ...ClockInstallOptions) (err error) {
	params := map[string]any{}
	if len(options) == 1 {
		if options[0].Time != nil {
			params, err = parseTime(options[0].Time)
			if err != nil {
				return err
			}
		}
	}

	_, err = c.browserCtx.channel.Send("clockInstall", params)

	return err
}

func (c *clockImpl) PauseAt(time any) error {
	params, err := parseTime(time)
	if err != nil {
		return err
	}

	_, err = c.browserCtx.channel.Send("clockPauseAt", params)
	return err
}

func (c *clockImpl) Resume() error {
	_, err := c.browserCtx.channel.Send("clockResume")
	return err
}

func (c *clockImpl) RunFor(ticks any) error {
	params, err := parseTicks(ticks)
	if err != nil {
		return err
	}

	_, err = c.browserCtx.channel.Send("clockRunFor", params)
	return err
}

func (c *clockImpl) SetFixedTime(time any) error {
	params, err := parseTime(time)
	if err != nil {
		return err
	}

	_, err = c.browserCtx.channel.Send("clockSetFixedTime", params)
	return err
}

func (c *clockImpl) SetSystemTime(time any) error {
	params, err := parseTime(time)
	if err != nil {
		return err
	}

	_, err = c.browserCtx.channel.Send("clockSetSystemTime", params)
	return err
}

func parseTime(t any) (map[string]any, error) {
	switch v := t.(type) {
	case int, int64:
		return map[string]any{"timeNumber": v}, nil
	case string:
		return map[string]any{"timeString": v}, nil
	case time.Time:
		return map[string]any{"timeNumber": v.UnixMilli()}, nil
	default:
		return nil, errors.New("time should be one of: int, int64, string, time.Time")
	}
}

func parseTicks(ticks any) (map[string]any, error) {
	switch v := ticks.(type) {
	case int, int64:
		return map[string]any{"ticksNumber": v}, nil
	case string:
		return map[string]any{"ticksString": v}, nil
	default:
		return nil, errors.New("ticks should be one of: int, int64, string")
	}
}
