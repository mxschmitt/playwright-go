package playwright

import (
	"fmt"
	"strings"

	"github.com/go-stack/stack"
)

type BindingCall interface {
	Call(f BindingCallFunction)
}

type bindingCallImpl struct {
	channelOwner
}

// BindingSource is the value passed to a binding call execution
type BindingSource struct {
	Context BrowserContext
	Page    Page
	Frame   Frame
}

// ExposedFunction represents the func signature of an exposed function
type ExposedFunction = func(args ...any) any

// BindingCallFunction represents the func signature of an exposed binding call func
type BindingCallFunction func(source *BindingSource, args ...any) any

func (b *bindingCallImpl) Call(f BindingCallFunction) {
	defer func() {
		if r := recover(); r != nil {
			// The recovered value may not be an error (e.g. panic("boom")); coerce
			// it so we always reject instead of panicking again on a failed type
			// assertion (which would crash the binding goroutine).
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			if _, sendErr := b.channel.Send("reject", map[string]any{
				"error": serializeError(err),
			}); sendErr != nil {
				logger.Error("could not reject BindingCall", "error", sendErr)
			}
		}
	}()

	frame := fromChannel(b.initializer["frame"]).(*frameImpl)
	source := &BindingSource{
		Context: frame.Page().Context(),
		Page:    frame.Page(),
		Frame:   frame,
	}
	var result any
	if handle, ok := b.initializer["handle"]; ok {
		result = f(source, fromChannel(handle))
	} else {
		initializerArgs := b.initializer["args"].([]any)
		funcArgs := []any{}
		for i := range initializerArgs {
			funcArgs = append(funcArgs, parseResult(initializerArgs[i]))
		}
		result = f(source, funcArgs...)
	}
	_, err := b.channel.Send("resolve", map[string]any{
		"result": serializeArgument(result),
	})
	if err != nil {
		logger.Error("could not resolve BindingCall", "error", err)
	}
}

func serializeError(err error) map[string]any {
	st := stack.Trace().TrimRuntime()
	if len(st) == 0 { // https://github.com/go-stack/stack/issues/27
		st = stack.Trace()
	}
	return map[string]any{
		"error": &Error{
			Name:    "Playwright for Go Error",
			Message: err.Error(),
			Stack: strings.ReplaceAll(strings.TrimFunc(fmt.Sprintf("%+v", st), func(r rune) bool {
				return r == '[' || r == ']'
			}), " ", "\n"),
		},
	}
}

func newBindingCall(parent *channelOwner, objectType string, guid string, initializer map[string]any) *bindingCallImpl {
	bt := &bindingCallImpl{}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	return bt
}
