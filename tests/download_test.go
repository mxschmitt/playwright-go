package playwright_test

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDownloadBasic(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/downloadWithFilename", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add("Content-Disposition", "attachment; filename=file.txt")
		if _, err := w.Write([]byte("foobar")); err != nil {
			log.Printf("could not write: %v", err)
		}
	})
	require.NoError(t, page.SetContent(
		fmt.Sprintf(`<a href="%s/downloadWithFilename">download</a>`, server.PREFIX),
	))

	download, err := page.ExpectDownload(func() error {
		return page.Locator("a").Click()
	})
	require.NoError(t, err)
	require.Equal(t, page, download.Page())
	require.Equal(t, download.URL(), fmt.Sprintf("%s/downloadWithFilename", server.PREFIX))
	require.Equal(t, download.SuggestedFilename(), "file.txt")
	require.Equal(t, download.String(), "file.txt")
	require.NoError(t, download.Failure())

	file, err := download.Path()
	require.NoError(t, err)
	require.FileExists(t, file)

	tmpFile := filepath.Join(t.TempDir(), download.SuggestedFilename())
	require.NoFileExists(t, tmpFile)
	require.NoError(t, download.SaveAs(tmpFile))
	require.FileExists(t, tmpFile)

	require.NoError(t, download.Delete())
	require.NoFileExists(t, file)
}

func TestDownloadCancel(t *testing.T) {
	BeforeEach(t)

	server.SetRoute("/downloadWithDelay", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add("Content-Disposition", "attachment")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				if _, err := w.Write([]byte("foobar")); err != nil {
					log.Printf("could not write: %v", err)
					return
				}
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	})
	require.NoError(t, page.SetContent(
		fmt.Sprintf(`<a href="%s/downloadWithDelay">download</a>`, server.PREFIX),
	))
	download, err := page.ExpectDownload(func() error {
		return page.Locator("a").Click()
	})
	require.NoError(t, err)
	require.NoError(t, download.Cancel())
	require.Error(t, download.Failure(), "canceled")
}
