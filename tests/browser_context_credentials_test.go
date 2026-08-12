package playwright_test

import (
	"path/filepath"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

const webAuthnRPID = "localhost"

const webAuthnAuthenticateScript = `async ({ rpId, credentialId }) => {
	const b64URLToBytes = value => {
		let padded = value.replace(/-/g, '+').replace(/_/g, '/');
		while (padded.length % 4)
			padded += '=';
		const binary = atob(padded);
		const bytes = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; ++i)
			bytes[i] = binary.charCodeAt(i);
		return bytes;
	};
	const credential = await navigator.credentials.get({
		publicKey: {
			challenge: crypto.getRandomValues(new Uint8Array(32)),
			rpId,
			allowCredentials: [{ type: 'public-key', id: b64URLToBytes(credentialId) }],
			userVerification: 'preferred',
		},
	});
	const response = credential.response;
	return {
		id: credential.id,
		type: credential.type,
		hasClientData: response.clientDataJSON.byteLength > 0,
		hasAuthData: response.authenticatorData.byteLength > 0,
		hasSignature: response.signature.byteLength > 0,
		hasUserPresentAndVerified: (new Uint8Array(response.authenticatorData)[32] & 0x05) === 0x05,
	};
}`

const webAuthnAuthenticateErrorScript = `async ({ rpId, credentialId }) => {
	const b64URLToBytes = value => {
		let padded = value.replace(/-/g, '+').replace(/_/g, '/');
		while (padded.length % 4)
			padded += '=';
		const binary = atob(padded);
		const bytes = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; ++i)
			bytes[i] = binary.charCodeAt(i);
		return bytes;
	};
	try {
		await navigator.credentials.get({
			publicKey: {
				challenge: crypto.getRandomValues(new Uint8Array(32)),
				rpId,
				allowCredentials: [{ type: 'public-key', id: b64URLToBytes(credentialId) }],
			},
		});
		return 'no-error';
	} catch (error) {
		return error.name;
	}
}`

const webAuthnCreateScript = `async ({ rpId }) => {
	const credential = await navigator.credentials.create({
		publicKey: {
			challenge: crypto.getRandomValues(new Uint8Array(32)),
			rp: { id: rpId, name: 'Test RP' },
			user: { id: new Uint8Array([1, 2, 3, 4]), name: 'u', displayName: 'User' },
			pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
			authenticatorSelection: { residentKey: 'required', userVerification: 'preferred' },
		},
	});
	return credential.id;
}`

const webAuthnDiscoverableCredentialScript = `async ({ rpId }) => {
	const credential = await navigator.credentials.get({
		publicKey: {
			challenge: crypto.getRandomValues(new Uint8Array(32)),
			rpId,
			userVerification: 'preferred',
		},
	});
	return credential.id;
}`

func newWebAuthnPage(t *testing.T, ctx playwright.BrowserContext) playwright.Page {
	t.Helper()
	p, err := ctx.NewPage()
	require.NoError(t, err)
	_, err = p.Goto(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	return p
}

func requireWebAuthnAuthentication(t *testing.T, p playwright.Page, rpID, credentialID string) {
	t.Helper()
	value, err := p.Evaluate(webAuthnAuthenticateScript, map[string]any{
		"rpId":         rpID,
		"credentialId": credentialID,
	})
	require.NoError(t, err)
	result, ok := value.(map[string]any)
	require.True(t, ok, "expected WebAuthn authentication result object, got %T", value)
	require.Equal(t, credentialID, result["id"])
	require.Equal(t, "public-key", result["type"])
	require.Equal(t, true, result["hasClientData"])
	require.Equal(t, true, result["hasAuthData"])
	require.Equal(t, true, result["hasSignature"])
	require.Equal(t, true, result["hasUserPresentAndVerified"])
}

func TestBrowserContextExposesCredentialsProperty(t *testing.T) {
	BeforeEach(t)

	require.NotNil(t, context.Credentials())
	// The same instance is returned on each access.
	require.Same(t, context.Credentials(), context.Credentials())
}

func TestBrowserContextInstallCreateGetDeleteCredentials(t *testing.T) {
	BeforeEach(t)

	// WebAuthn requires a secure context; the test server's localhost origin qualifies.
	_, err := page.Goto(server.CROSS_PROCESS_PREFIX+"/empty.html", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	creds := context.Credentials()
	require.NoError(t, creds.Install())

	created, err := creds.Create("localhost")
	require.NoError(t, err)
	require.Equal(t, "localhost", created.RpId)
	require.NotEmpty(t, created.Id)

	list, err := creds.Get()
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, created.Id, list[0].Id)

	require.NoError(t, creds.Delete(created.Id))
	list, err = creds.Get()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestBrowserContextCredentialsDoNotInterceptBeforeInstall(t *testing.T) {
	BeforeEach(t)

	_, err := context.Credentials().Create(webAuthnRPID)
	require.NoError(t, err)
	_, err = page.Goto(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	installed, err := page.Evaluate(`() => globalThis.__pwWebAuthnInstalled === true`)
	require.NoError(t, err)
	require.Equal(t, false, installed)
}

func TestBrowserContextCredentialsSeedKnownCredentialAndAuthenticate(t *testing.T) {
	BeforeEach(t)

	source, err := browser.NewContext()
	require.NoError(t, err)
	defer source.Close() //nolint:errcheck
	known, err := source.Credentials().Create(webAuthnRPID)
	require.NoError(t, err)

	target, err := browser.NewContext()
	require.NoError(t, err)
	defer target.Close() //nolint:errcheck
	_, err = target.Credentials().Create(known.RpId, playwright.CredentialsCreateOptions{
		Id:         playwright.String(known.Id),
		UserHandle: playwright.String(known.UserHandle),
		PrivateKey: playwright.String(known.PrivateKey),
		PublicKey:  playwright.String(known.PublicKey),
	})
	require.NoError(t, err)
	require.NoError(t, target.Credentials().Install())
	targetPage := newWebAuthnPage(t, target)
	requireWebAuthnAuthentication(t, targetPage, known.RpId, known.Id)

	require.NoError(t, target.Credentials().Delete(known.Id))
	credentials, err := target.Credentials().Get()
	require.NoError(t, err)
	require.Empty(t, credentials)
	failure, err := targetPage.Evaluate(webAuthnAuthenticateErrorScript, map[string]any{
		"rpId":         known.RpId,
		"credentialId": known.Id,
	})
	require.NoError(t, err)
	require.Equal(t, "NotAllowedError", failure)
}

func TestBrowserContextCredentialsCapturePageCredentialAndReuseIt(t *testing.T) {
	BeforeEach(t)

	setupContext, err := browser.NewContext()
	require.NoError(t, err)
	defer setupContext.Close() //nolint:errcheck
	require.NoError(t, setupContext.Credentials().Install())
	setupPage := newWebAuthnPage(t, setupContext)
	createdID, err := setupPage.Evaluate(webAuthnCreateScript, map[string]any{"rpId": webAuthnRPID})
	require.NoError(t, err)
	createdIDString, ok := createdID.(string)
	require.True(t, ok, "expected a base64url credential id, got %T", createdID)

	captured, err := setupContext.Credentials().Get(playwright.CredentialsGetOptions{RpId: playwright.String(webAuthnRPID)})
	require.NoError(t, err)
	require.Len(t, captured, 1)
	require.Equal(t, createdIDString, captured[0].Id)
	require.Regexp(t, `^[A-Za-z0-9_-]+$`, captured[0].PrivateKey)
	require.Regexp(t, `^[A-Za-z0-9_-]+$`, captured[0].PublicKey)

	target, err := browser.NewContext()
	require.NoError(t, err)
	defer target.Close() //nolint:errcheck
	_, err = target.Credentials().Create(captured[0].RpId, playwright.CredentialsCreateOptions{
		Id:         playwright.String(captured[0].Id),
		UserHandle: playwright.String(captured[0].UserHandle),
		PrivateKey: playwright.String(captured[0].PrivateKey),
		PublicKey:  playwright.String(captured[0].PublicKey),
	})
	require.NoError(t, err)
	require.NoError(t, target.Credentials().Install())
	targetPage := newWebAuthnPage(t, target)
	gotID, err := targetPage.Evaluate(webAuthnDiscoverableCredentialScript, map[string]any{"rpId": webAuthnRPID})
	require.NoError(t, err)
	require.Equal(t, createdIDString, gotID)
}

func TestBrowserContextCredentialsStorageStateReusesPageCredential(t *testing.T) {
	BeforeEach(t)

	setupContext, err := browser.NewContext()
	require.NoError(t, err)
	defer setupContext.Close() //nolint:errcheck
	require.NoError(t, setupContext.Credentials().Install())
	setupPage := newWebAuthnPage(t, setupContext)
	createdID, err := setupPage.Evaluate(webAuthnCreateScript, map[string]any{"rpId": webAuthnRPID})
	require.NoError(t, err)
	createdIDString, ok := createdID.(string)
	require.True(t, ok, "expected a base64url credential id, got %T", createdID)

	state, err := setupContext.StorageState(playwright.BrowserContextStorageStateOptions{
		Credentials: playwright.Bool(true),
	})
	require.NoError(t, err)
	require.Len(t, state.Credentials, 1)
	require.Equal(t, createdIDString, state.Credentials[0].Id)

	restored, err := browser.NewContext(playwright.BrowserNewContextOptions{
		StorageState: state.ToOptionalStorageState(),
	})
	require.NoError(t, err)
	defer restored.Close() //nolint:errcheck
	restoredPage := newWebAuthnPage(t, restored)
	gotID, err := restoredPage.Evaluate(webAuthnDiscoverableCredentialScript, map[string]any{"rpId": webAuthnRPID})
	require.NoError(t, err)
	require.Equal(t, createdIDString, gotID)
}

func TestStorageStateRoundTripWebAuthnCredentials(t *testing.T) {
	BeforeEach(t)
	require.NoError(t, context.AddCookies([]playwright.OptionalCookie{{
		Name:  "session",
		Value: "cookie-value",
		URL:   playwright.String(server.PREFIX),
	}}))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => localStorage.setItem("roll-key", "roll-value")`)
	require.NoError(t, err)

	// Seed a virtual credential.
	require.NoError(t, context.Credentials().Install())
	cred, err := context.Credentials().Create("example.com")
	require.NoError(t, err)
	require.NotEmpty(t, cred.Id)
	withoutCredentials, err := context.StorageState()
	require.NoError(t, err)
	require.Empty(t, withoutCredentials.Credentials, "credentials must be opt-in")
	require.NotEmpty(t, withoutCredentials.Cookies)
	require.NotEmpty(t, withoutCredentials.Origins)

	state, err := context.StorageState(playwright.BrowserContextStorageStateOptions{
		Credentials: playwright.Bool(true),
	})
	require.NoError(t, err)
	require.NotEmpty(t, state.Credentials)
	require.Equal(t, cred.Id, state.Credentials[0].Id)

	// In-memory round-trip via OptionalStorageState.
	opt := state.ToOptionalStorageState()
	require.NotEmpty(t, opt.Credentials)

	ctx2, err := browser.NewContext(playwright.BrowserNewContextOptions{
		StorageState: opt,
	})
	require.NoError(t, err)
	defer ctx2.Close() //nolint:errcheck
	restored, err := ctx2.Credentials().Get()
	require.NoError(t, err)
	require.NotEmpty(t, restored)
	require.Equal(t, cred.Id, restored[0].Id)

	// Path round-trip.
	path := filepath.Join(t.TempDir(), "state.json")
	_, err = context.StorageState(playwright.BrowserContextStorageStateOptions{
		Credentials: playwright.Bool(true),
		Path:        playwright.String(path),
	})
	require.NoError(t, err)
	ctx3, err := browser.NewContext(playwright.BrowserNewContextOptions{
		StorageStatePath: playwright.String(path),
	})
	require.NoError(t, err)
	defer ctx3.Close() //nolint:errcheck
	restored2, err := ctx3.Credentials().Get()
	require.NoError(t, err)
	require.NotEmpty(t, restored2)

	// SetStorageState must restore and subsequently clear credentials.
	ctx4, err := browser.NewContext()
	require.NoError(t, err)
	defer ctx4.Close() //nolint:errcheck
	require.NoError(t, ctx4.SetStorageState(path))
	restored3, err := ctx4.Credentials().Get()
	require.NoError(t, err)
	require.NotEmpty(t, restored3)
	withoutCredentialsPath := filepath.Join(t.TempDir(), "state-without-credentials.json")
	_, err = context.StorageState(playwright.BrowserContextStorageStateOptions{
		Path: playwright.String(withoutCredentialsPath),
	})
	require.NoError(t, err)
	require.NoError(t, ctx4.SetStorageState(withoutCredentialsPath))
	restored3, err = ctx4.Credentials().Get()
	require.NoError(t, err)
	require.Empty(t, restored3)

	// APIRequestContext strips credentials while preserving cookies/origins.
	req, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		StorageState: state,
	})
	require.NoError(t, err)
	reqState, err := req.StorageState()
	require.NoError(t, err)
	require.Empty(t, reqState.Credentials)
	require.Equal(t, state.Cookies, reqState.Cookies)
	require.Equal(t, state.Origins, reqState.Origins)
	require.NoError(t, req.Dispose())

	reqFromFile, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		StorageStatePath: playwright.String(path),
	})
	require.NoError(t, err)
	reqFileState, err := reqFromFile.StorageState()
	require.NoError(t, err)
	require.Empty(t, reqFileState.Credentials)
	require.Equal(t, state.Cookies, reqFileState.Cookies)
	require.Equal(t, state.Origins, reqFileState.Origins)
	require.NoError(t, reqFromFile.Dispose())
}
