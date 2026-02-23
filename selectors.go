package playwright

import (
	"errors"
	"os"
	"sync"
)

type selectorsOwnerImpl struct {
	channelOwner
}

func (s *selectorsOwnerImpl) setTestIdAttributeName(name string) {
	s.channel.SendNoReply("setTestIdAttributeName", map[string]any{
		"testIdAttributeName": name,
	})
}

func newSelectorsOwner(parent *channelOwner, objectType string, guid string, initializer map[string]any) *selectorsOwnerImpl {
	obj := &selectorsOwnerImpl{}
	obj.createChannelOwner(obj, parent, objectType, guid, initializer)
	return obj
}

type selectorsImpl struct {
	mu            sync.RWMutex // protects registrations slice
	contexts      sync.Map     // map of BrowserContext channels
	registrations []map[string]any
}

func (s *selectorsImpl) Register(name string, script Script, options ...SelectorsRegisterOptions) error {
	if script.Path == nil && script.Content == nil {
		return errors.New("Either source or path should be specified")
	}
	source := ""
	if script.Path != nil {
		content, err := os.ReadFile(*script.Path)
		if err != nil {
			return err
		}
		source = string(content)
	} else {
		source = *script.Content
	}
	selectorEngine := map[string]any{
		"name":   name,
		"source": source,
	}
	if len(options) == 1 && options[0].ContentScript != nil {
		selectorEngine["contentScript"] = *options[0].ContentScript
	}
	params := map[string]any{
		"selectorEngine": selectorEngine,
	}
	// Register with all active contexts, ignoring contexts that have been closed
	s.contexts.Range(func(key, value any) bool {
		_, _ = value.(*browserContextImpl).channel.Send("registerSelectorEngine", params)
		// Continue to next context even if this one failed (e.g., context closed)
		return true
	})
	s.mu.Lock()
	s.registrations = append(s.registrations, selectorEngine)
	s.mu.Unlock()
	return nil
}

func (s *selectorsImpl) SetTestIdAttribute(name string) {
	setTestIdAttributeName(name)
	s.contexts.Range(func(key, value any) bool {
		value.(*browserContextImpl).channel.SendNoReply("setTestIdAttributeName", map[string]any{
			"testIdAttributeName": name,
		})
		return true
	})
}

func (s *selectorsImpl) addChannel(channel *selectorsOwnerImpl) {
	// Legacy support for older Playwright versions with server-side selectors
	s.contexts.Store(channel.guid, channel)
	s.mu.RLock()
	for _, selectorEngine := range s.registrations {
		params := map[string]any{
			"selectorEngine": selectorEngine,
		}
		channel.channel.SendNoReply("registerSelectorEngine", params)
		channel.setTestIdAttributeName(getTestIdAttributeName())
	}
	s.mu.RUnlock()
}

func (s *selectorsImpl) removeChannel(channel *selectorsOwnerImpl) {
	// Legacy support for older Playwright versions with server-side selectors
	s.contexts.Delete(channel.guid)
}

func (s *selectorsImpl) addContext(context *browserContextImpl) {
	s.contexts.Store(context.guid, context)
	s.mu.RLock()
	for _, selectorEngine := range s.registrations {
		params := map[string]any{
			"selectorEngine": selectorEngine,
		}
		context.channel.SendNoReply("registerSelectorEngine", params)
	}
	s.mu.RUnlock()
	testIdAttr := getTestIdAttributeName()
	if testIdAttr != "" {
		context.channel.SendNoReply("setTestIdAttributeName", map[string]any{
			"testIdAttributeName": testIdAttr,
		})
	}
}

func (s *selectorsImpl) removeContext(context *browserContextImpl) {
	s.contexts.Delete(context.guid)
}

func newSelectorsImpl() *selectorsImpl {
	return &selectorsImpl{
		contexts:      sync.Map{},
		registrations: make([]map[string]any, 0),
	}
}
