package playwright

type webStorageImpl struct {
	page *pageImpl
	kind string // "local" or "session"
}

func newWebStorage(page *pageImpl, kind string) *webStorageImpl {
	return &webStorageImpl{page: page, kind: kind}
}

func (w *webStorageImpl) Items() ([]WebStorageItem, error) {
	result, err := w.page.channel.SendReturnAsDict("webStorageItems", map[string]any{"kind": w.kind})
	if err != nil {
		return nil, err
	}
	items := make([]WebStorageItem, 0)
	if rawItems, ok := result["items"].([]any); ok {
		for _, raw := range rawItems {
			if item, ok := raw.(map[string]any); ok {
				items = append(items, WebStorageItem{
					Name:  item["name"].(string),
					Value: item["value"].(string),
				})
			}
		}
	}
	return items, nil
}

func (w *webStorageImpl) GetItem(name string) (string, error) {
	result, err := w.page.channel.SendReturnAsDict("webStorageGetItem", map[string]any{"kind": w.kind, "name": name})
	if err != nil {
		return "", err
	}
	if value, ok := result["value"].(string); ok {
		return value, nil
	}
	return "", nil
}

func (w *webStorageImpl) SetItem(name string, value string) error {
	_, err := w.page.channel.Send("webStorageSetItem", map[string]any{"kind": w.kind, "name": name, "value": value})
	return err
}

func (w *webStorageImpl) RemoveItem(name string) error {
	_, err := w.page.channel.Send("webStorageRemoveItem", map[string]any{"kind": w.kind, "name": name})
	return err
}

func (w *webStorageImpl) Clear() error {
	_, err := w.page.channel.Send("webStorageClear", map[string]any{"kind": w.kind})
	return err
}
