package playwright

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-stack/stack"
	"github.com/playwright-community/playwright-go/internal/safe"
)

var (
	pkgSourcePathPattern = regexp.MustCompile(`.+[\\/]playwright-go[\\/][^\\/]+\.go`)
	apiNameTransform     = regexp.MustCompile(`(?U)\(\*(.+)(Impl)?\)`)
)

type connection struct {
	transport    transport
	apiZone      sync.Map
	objects      *safe.SyncMap[string, *channelOwner]
	lastID       atomic.Uint32
	rootObject   *rootChannelOwner
	callbacks    *safe.SyncMap[uint32, *protocolCallback]
	afterClose   func()
	onClose      func() error
	isRemote     bool
	localUtils   *localUtilsImpl
	tracingCount atomic.Int32
	abort        chan struct{}
	abortOnce    sync.Once
	err          *safeValue[error] // for event listener error
	closedError  *safeValue[error]
}

func (c *connection) Start() (*Playwright, error) {
	go func() {
		for {
			msg, err := c.transport.Poll()
			if err != nil {
				_ = c.transport.Close()
				c.cleanup(err)
				return
			}
			c.Dispatch(msg)
		}
	}()

	c.onClose = func() error {
		if err := c.transport.Close(); err != nil {
			return err
		}
		return nil
	}

	return c.rootObject.initialize()
}

func (c *connection) Stop() error {
	if err := c.onClose(); err != nil {
		return err
	}
	c.cleanup()
	return nil
}

func (c *connection) cleanup(cause ...error) {
	if len(cause) > 0 {
		c.closedError.Set(fmt.Errorf("%w: %w", ErrTargetClosed, cause[0]))
	} else {
		c.closedError.Set(ErrTargetClosed)
	}
	if c.afterClose != nil {
		c.afterClose()
	}
	c.abortOnce.Do(func() {
		select {
		case <-c.abort:
		default:
			close(c.abort)
		}
	})
}

func (c *connection) Dispatch(msg *message) {
	if c.closedError.Get() != nil {
		return
	}
	method := msg.Method
	if msg.ID != 0 {
		cb, _ := c.callbacks.LoadAndDelete(uint32(msg.ID))
		if cb.noReply {
			return
		}
		if msg.Error != nil {
			cb.SetError(parseError(msg.Error.Error))
		} else {
			// Always resolve GUIDs in responses, regardless of connection type
			// The protocol guarantees that __create__ events arrive before responses that reference those objects
			result, err := c.replaceGuidsWithChannels(msg.Result)
			if err != nil {
				cb.SetError(fmt.Errorf("failed to resolve response objects: %w", err))
			} else {
				cb.SetResult(result.(map[string]any))
			}
		}
		return
	}
	object, _ := c.objects.Load(msg.GUID)
	if method == "__create__" {
		_, err := c.createRemoteObject(
			object, msg.Params["type"].(string), msg.Params["guid"].(string), msg.Params["initializer"],
		)
		if err != nil {
			// Critical: object creation failure indicates corrupted protocol state
			// Close connection to prevent cascade failures
			c.cleanup(err)
		}
		return
	}
	if object == nil {
		return
	}
	if method == "__adopt__" {
		child, ok := c.objects.Load(msg.Params["guid"].(string))
		if !ok {
			return
		}
		object.adopt(child)
		return
	}
	if method == "__dispose__" {
		reason, ok := msg.Params["reason"]
		if ok {
			object.dispose(reason.(string))
		} else {
			object.dispose()
		}
		return
	}
	if object.objectType == "JsonPipe" {
		object.channel.Emit(method, msg.Params)
	} else {
		// Always resolve GUIDs in events, regardless of connection type
		// The protocol guarantees that __create__ events arrive before events that reference those objects
		params, err := c.replaceGuidsWithChannels(msg.Params)
		if err != nil {
			// Event parameters contain invalid references - connection is corrupted
			c.cleanup(fmt.Errorf("failed to resolve event parameters for %s: %w", method, err))
			return
		}
		object.channel.Emit(method, params)
	}
}

func (c *connection) LocalUtils() *localUtilsImpl {
	return c.localUtils
}

func (c *connection) createRemoteObject(parent *channelOwner, objectType string, guid string, initializer any) (any, error) {
	resolved, err := c.replaceGuidsWithChannels(initializer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve initializer for %s (guid=%s): %w", objectType, guid, err)
	}
	result := createObjectFactory(parent, objectType, guid, resolved.(map[string]any))
	return result, nil
}

func (c *connection) WrapAPICall(cb func() (any, error), isInternal bool) (any, error) {
	if _, ok := c.apiZone.Load("apiZone"); ok {
		return cb()
	}
	c.apiZone.Store("apiZone", serializeCallStack(isInternal))
	return cb()
}

func (c *connection) replaceGuidsWithChannels(payload any) (any, error) {
	if payload == nil {
		return nil, nil
	}
	v := reflect.ValueOf(payload)
	if v.Kind() == reflect.Slice {
		listV := payload.([]any)
		for i := range listV {
			resolved, err := c.replaceGuidsWithChannels(listV[i])
			if err != nil {
				return nil, fmt.Errorf("failed to resolve slice element at index %d: %w", i, err)
			}
			listV[i] = resolved
		}
		return listV, nil
	}
	if v.Kind() == reflect.Map {
		mapV := payload.(map[string]any)
		// Check if this map represents an object reference (has "guid" field)
		if guid, hasGUID := mapV["guid"]; hasGUID {
			guidStr, ok := guid.(string)
			if !ok {
				return nil, fmt.Errorf("guid field is not a string: %T", guid)
			}
			// Try to load the object from connection's objects map
			if channelOwner, ok := c.objects.Load(guidStr); ok {
				return channelOwner.channel, nil
			}
			// Object not found - this indicates a protocol error or message ordering issue
			return nil, fmt.Errorf("object with guid %s was not bound in the connection", guidStr)
		}
		// Recursively process all values in the map
		for key := range mapV {
			resolved, err := c.replaceGuidsWithChannels(mapV[key])
			if err != nil {
				return nil, fmt.Errorf("failed to resolve map key '%s': %w", key, err)
			}
			mapV[key] = resolved
		}
		return mapV, nil
	}
	return payload, nil
}

func (c *connection) sendMessageToServer(object *channelOwner, method string, params any, noReply bool) (cb *protocolCallback) {
	cb = newProtocolCallback(noReply, c.abort)

	if err := c.closedError.Get(); err != nil {
		cb.SetError(err)
		return
	}
	if object.wasCollected {
		cb.SetError(errors.New("The object has been collected to prevent unbounded heap growth."))
		return
	}

	id := c.lastID.Add(1)
	c.callbacks.Store(id, cb)
	var (
		metadata = make(map[string]any, 0)
		stack    = make([]map[string]any, 0)
	)
	apiZone, ok := c.apiZone.LoadAndDelete("apiZone")
	if ok {
		for k, v := range apiZone.(parsedStackTrace).metadata {
			metadata[k] = v
		}
		stack = append(stack, apiZone.(parsedStackTrace).frames...)
	}
	metadata["wallTime"] = time.Now().UnixMilli()
	message := map[string]any{
		"id":       id,
		"guid":     object.guid,
		"method":   method,
		"params":   params, // channel.MarshalJSON will replace channel with guid
		"metadata": metadata,
	}
	if c.tracingCount.Load() > 0 && len(stack) > 0 && object.guid != "localUtils" {
		c.LocalUtils().AddStackToTracingNoReply(id, stack)
	}

	if err := c.transport.Send(message); err != nil {
		cb.SetError(fmt.Errorf("could not send message: %w", err))
		return
	}

	return
}

func (c *connection) setInTracing(isTracing bool) {
	if isTracing {
		c.tracingCount.Add(1)
	} else {
		c.tracingCount.Add(-1)
	}
}

type parsedStackTrace struct {
	frames   []map[string]any
	metadata map[string]any
}

func serializeCallStack(isInternal bool) parsedStackTrace {
	st := stack.Trace().TrimRuntime()
	if len(st) == 0 { // https://github.com/go-stack/stack/issues/27
		st = stack.Trace()
	}

	lastInternalIndex := 0
	for i, s := range st {
		if pkgSourcePathPattern.MatchString(s.Frame().File) {
			lastInternalIndex = i
		}
	}
	apiName := ""
	if !isInternal {
		apiName = fmt.Sprintf("%n", st[lastInternalIndex])
	}
	st = st.TrimBelow(st[lastInternalIndex])

	callStack := make([]map[string]any, 0)
	for i, s := range st {
		if i == 0 {
			continue
		}
		callStack = append(callStack, map[string]any{
			"file":     s.Frame().File,
			"line":     s.Frame().Line,
			"column":   0,
			"function": s.Frame().Function,
		})
	}
	metadata := make(map[string]any)
	if len(st) > 1 {
		metadata["location"] = serializeCallLocation(st[1])
	}
	apiName = apiNameTransform.ReplaceAllString(apiName, "$1")
	if len(apiName) > 1 {
		apiName = strings.ToUpper(apiName[:1]) + apiName[1:]
	}
	metadata["apiName"] = apiName
	metadata["isInternal"] = isInternal
	return parsedStackTrace{
		metadata: metadata,
		frames:   callStack,
	}
}

func serializeCallLocation(caller stack.Call) map[string]any {
	line, _ := strconv.Atoi(fmt.Sprintf("%d", caller))
	return map[string]any{
		"file": fmt.Sprintf("%s", caller),
		"line": line,
	}
}

func newConnection(transport transport, localUtils ...*localUtilsImpl) *connection {
	connection := &connection{
		abort:       make(chan struct{}, 1),
		callbacks:   safe.NewSyncMap[uint32, *protocolCallback](),
		objects:     safe.NewSyncMap[string, *channelOwner](),
		transport:   transport,
		isRemote:    false,
		err:         &safeValue[error]{},
		closedError: &safeValue[error]{},
	}
	if len(localUtils) > 0 {
		connection.localUtils = localUtils[0]
		connection.isRemote = true
	}
	connection.rootObject = newRootChannelOwner(connection)
	return connection
}

func fromChannel(v any) any {
	if ch, ok := v.(*channel); ok {
		return ch.object
	}
	panic(fmt.Sprintf("fromChannel: expected *channel, got %T: %+v", v, v))
}

func fromNullableChannel(v any) any {
	if v == nil {
		return nil
	}
	return fromChannel(v)
}

type protocolCallback struct {
	done    chan struct{}
	noReply bool
	abort   <-chan struct{}
	once    sync.Once
	value   map[string]any
	err     error
}

func (pc *protocolCallback) setResultOnce(result map[string]any, err error) {
	pc.once.Do(func() {
		pc.value = result
		pc.err = err
		if pc.done != nil {
			close(pc.done)
		}
	})
}

func (pc *protocolCallback) waitResult() {
	if pc.noReply {
		return
	}
	select {
	case <-pc.done: // wait for result
		return
	case <-pc.abort:
		select {
		case <-pc.done:
			return
		default:
			pc.err = errors.New("Connection closed")
			return
		}
	}
}

func (pc *protocolCallback) SetError(err error) {
	pc.setResultOnce(nil, err)
}

func (pc *protocolCallback) SetResult(result map[string]any) {
	pc.setResultOnce(result, nil)
}

func (pc *protocolCallback) GetResult() (map[string]any, error) {
	pc.waitResult()
	return pc.value, pc.err
}

// GetResultValue returns value if the map has only one element
func (pc *protocolCallback) GetResultValue() (any, error) {
	pc.waitResult()
	if len(pc.value) == 0 { // empty map treated as nil
		return nil, pc.err
	}
	if len(pc.value) == 1 {
		for key := range pc.value {
			return pc.value[key], pc.err
		}
	}

	return pc.value, pc.err
}

func newProtocolCallback(noReply bool, abort <-chan struct{}) *protocolCallback {
	if noReply {
		return &protocolCallback{
			noReply: true,
			abort:   abort,
		}
	}
	return &protocolCallback{
		done:  make(chan struct{}, 1),
		abort: abort,
	}
}
