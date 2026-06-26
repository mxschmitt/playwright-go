package playwright

type credentialsImpl struct {
	browserCtx *browserContextImpl
}

func newCredentials(bCtx *browserContextImpl) Credentials {
	return &credentialsImpl{browserCtx: bCtx}
}

func (c *credentialsImpl) Install() error {
	_, err := c.browserCtx.channel.Send("credentialsInstall")
	return err
}

func (c *credentialsImpl) Create(rpId string, options ...CredentialsCreateOptions) (*VirtualCredential, error) {
	overrides := map[string]any{"rpId": rpId}
	result, err := c.browserCtx.channel.SendReturnAsDict("credentialsCreate", options, overrides)
	if err != nil {
		return nil, err
	}
	credential := &VirtualCredential{}
	remapMapToStruct(result["credential"], credential)
	return credential, nil
}

func (c *credentialsImpl) Delete(id string) error {
	_, err := c.browserCtx.channel.Send("credentialsDelete", map[string]any{"id": id})
	return err
}

func (c *credentialsImpl) Get(options ...CredentialsGetOptions) ([]VirtualCredential, error) {
	result, err := c.browserCtx.channel.SendReturnAsDict("credentialsGet", options)
	if err != nil {
		return nil, err
	}
	credentials := make([]VirtualCredential, 0)
	if rawCredentials, ok := result["credentials"].([]any); ok {
		for _, raw := range rawCredentials {
			credential := VirtualCredential{}
			remapMapToStruct(raw, &credential)
			credentials = append(credentials, credential)
		}
	}
	return credentials, nil
}
