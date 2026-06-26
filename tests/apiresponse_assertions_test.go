package playwright_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssertionsResponseIsOKPass(t *testing.T) {
	BeforeEach(t)

	response, err := page.Request().Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NoError(t, expect.APIResponse(response).ToBeOK())
	require.Error(t, expect.APIResponse(response).Not().ToBeOK())
}

func TestAssertionsShouldPrintResponseWithTextContentTypeIfToBeOKFails(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/text-content-type", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Text error"))
	})
	server.SetRoute("/no-content-type", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("No content type error"))
	})
	server.SetRoute("/binary-content-type", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/bmp")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Image content type error"))
	})

	response, err := page.Request().Get(server.PREFIX + "/text-content-type")
	require.NoError(t, err)
	require.NoError(t, expect.APIResponse(response).Not().ToBeOK())
	err = expect.APIResponse(response).ToBeOK()
	require.ErrorContains(t, err, "→ GET "+server.PREFIX+"/text-content-type")
	require.ErrorContains(t, err, "← 404 Not Found")
	require.ErrorContains(t, err, "Response text:")
	require.ErrorContains(t, err, "Text error")

	response, err = page.Request().Get(server.PREFIX + "/no-content-type")
	require.NoError(t, err)
	require.NoError(t, expect.APIResponse(response).Not().ToBeOK())
	err = expect.APIResponse(response).ToBeOK()
	require.ErrorContains(t, err, "→ GET "+server.PREFIX+"/no-content-type")
	require.ErrorContains(t, err, "← 404 Not Found")
	require.NotContains(t, err.Error(), "Response text:")
	require.NotContains(t, err.Error(), "No content type error")

	response, err = page.Request().Get(server.PREFIX + "/binary-content-type")
	require.NoError(t, err)
	err = expect.APIResponse(response).ToBeOK()
	require.ErrorContains(t, err, "→ GET "+server.PREFIX+"/binary-content-type")
	require.ErrorContains(t, err, "← 404 Not Found")
	require.NotContains(t, err.Error(), "Response text:")
	require.NotContains(t, err.Error(), "Image content type error")
}
