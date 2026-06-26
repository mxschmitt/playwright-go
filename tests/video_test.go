package playwright_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/h2non/filetype"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoShouldWork(t *testing.T) {
	recordVideoDir := t.TempDir()
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: playwright.String(recordVideoDir),
			Size: &playwright.Size{
				Width:  500,
				Height: 400,
			},
		},
	})

	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	_, err = page.Reload()
	require.NoError(t, err)
	_, err = page.Reload()
	require.NoError(t, err)
	//nolint:staticcheck
	page.WaitForTimeout(500) // make sure video has some data
	require.NoError(t, context.Close())

	path, err := page.Video().Path()
	require.NoError(t, err)
	files, err := os.ReadDir(recordVideoDir)
	require.NoError(t, err)
	require.Equal(t, len(files), 1)
	videoFileLocation := filepath.Join(recordVideoDir, files[0].Name())
	require.Equal(t, videoFileLocation, path)
	require.FileExists(t, videoFileLocation)
	content, err := os.ReadFile(videoFileLocation)
	require.NoError(t, err)
	require.True(t, filetype.IsVideo(content))
	tmpFile := filepath.Join(t.TempDir(), "test.webm")
	require.NoError(t, page.Video().SaveAs(tmpFile))
	require.FileExists(t, tmpFile)
	require.NoError(t, page.Video().Delete())
	require.NoFileExists(t, videoFileLocation)
}

func TestVideo(t *testing.T) {
	t.Run("should expose video path", func(t *testing.T) {
		recordVideoDir := t.TempDir()
		BeforeEach(t, playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(recordVideoDir),
				Size: &playwright.Size{
					Width:  500,
					Height: 400,
				},
			},
		})

		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, recordVideoDir)
		//nolint:staticcheck
		page.WaitForTimeout(500)
		require.NoError(t, page.Context().Close())
	})

	t.Run("should work when access video after close page", func(t *testing.T) {
		recordVideoDir := t.TempDir()
		BeforeEach(t, playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(recordVideoDir),
				Size: &playwright.Size{
					Width:  500,
					Height: 400,
				},
			},
		})

		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		//nolint:staticcheck
		page.WaitForTimeout(500)
		require.NoError(t, page.Close())
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, recordVideoDir)
		require.EventuallyWithT(t, func(collect *assert.CollectT) {
			require.FileExists(collect, path)
		}, 1*time.Second, 10*time.Millisecond)
	})

	t.Run("video should not exist when delete before close page", func(t *testing.T) {
		recordVideoDir := t.TempDir()
		BeforeEach(t, playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(recordVideoDir),
				Size: &playwright.Size{
					Width:  500,
					Height: 400,
				},
			},
		})

		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		//nolint:staticcheck
		page.WaitForTimeout(500)
		require.NoError(t, page.Close())
		require.NoError(t, video.Delete())
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, recordVideoDir)
		require.NoFileExists(t, path)
	})

	t.Run("video should not exist when no dir specified", func(t *testing.T) {
		BeforeEach(t)

		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.Error(t, err)
		require.Empty(t, path)
		tmpFile := filepath.Join(t.TempDir(), "test.webm")
		require.Error(t, video.SaveAs(tmpFile))
		require.NoError(t, page.Context().Close())
		require.Error(t, video.SaveAs(tmpFile))
		require.NoError(t, video.Delete())
	})

	t.Run("record video to path persistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		BeforeEach(t)

		require.NoError(t, context.Close())

		bt := browser.BrowserType()

		context, err := bt.LaunchPersistentContext(tmpDir, playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless: playwright.Bool(os.Getenv("HEADFUL") == ""),
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(tmpDir),
			},
		})
		require.NoError(t, err)
		page := context.Pages()[0]
		_, err = page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, tmpDir)
		//nolint:staticcheck
		page.WaitForTimeout(500)
		require.NoError(t, context.Close())
		require.EventuallyWithT(t, func(collect *assert.CollectT) {
			require.FileExists(collect, path)
		}, 1*time.Second, 10*time.Millisecond)
	})

	t.Run("remote server should work with saveas", func(t *testing.T) {
		tmpDir := t.TempDir()
		BeforeEach(t)

		remoteServer, err := newRemoteServer()
		require.NoError(t, err)
		defer remoteServer.Close()

		browser1, err := browserType.Connect(remoteServer.url)
		require.NoError(t, err)
		require.NotNil(t, browser1)
		defer browser1.Close() //nolint:errcheck

		browser_context, err := browser1.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(tmpDir),
			},
		})
		require.NoError(t, err)
		page, err = browser_context.NewPage()
		require.NoError(t, err)
		_, err = page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		//nolint:staticcheck
		page.WaitForTimeout(500)
		video := page.Video()
		_, err = video.Path()
		require.ErrorContains(t, err, "Path is not available when connecting remotely")
		tmpFile := filepath.Join(t.TempDir(), "test.webm")
		require.ErrorContains(t, video.SaveAs(tmpFile), "Page is not yet closed.")
		require.NoError(t, browser_context.Close())
		require.NoError(t, video.SaveAs(tmpFile))
		require.FileExists(t, tmpFile)
	})
}

func TestVideoRelativeDirShouldResolveToAbsolute(t *testing.T) {
	// Regression test for https://github.com/mxschmitt/playwright-go/issues/565
	// A relative recordVideo.dir must be resolved to an absolute path client-side
	// (matching upstream path.resolve), so Video().Path() does not depend on the
	// process working directory at the time it is read.
	//
	// Run from inside a temp dir so the relative path is valid on Windows too,
	// where t.TempDir() and the repo may live on different drives (and thus have
	// no relative path between them).
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { require.NoError(t, os.Chdir(origWd)) }()

	const relDir = "videos"
	// t.TempDir() may resolve through a symlink (e.g. /var -> /private/var on
	// macOS); resolve the expected absolute dir the same way the client does.
	expectedDir, err := filepath.Abs(relDir)
	require.NoError(t, err)

	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: playwright.String(relDir),
		},
	})

	_, err = page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	//nolint:staticcheck
	page.WaitForTimeout(500)
	require.NoError(t, context.Close())

	path, err := page.Video().Path()
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(path), "Video().Path() should be absolute, got %q", path)
	require.Equal(t, expectedDir, filepath.Dir(path))
	require.FileExists(t, path)
}

func TestScreencastStartStop(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)

	screencast, err := page.Screencast()
	require.NoError(t, err)

	var (
		mu         sync.Mutex
		frames     [][]byte
		firstFrame = make(chan struct{}, 1)
	)
	err = screencast.Start(playwright.ScreencastStartOptions{
		OnFrame: func(frame playwright.OnFrame) {
			mu.Lock()
			frames = append(frames, frame.Data)
			mu.Unlock()
			select {
			case firstFrame <- struct{}{}:
			default:
			}
		},
	})
	require.NoError(t, err)

	// Trigger some activity to generate frames, then wait until at least one
	// frame arrives rather than racing a fixed timeout (webkit can be slow).
	_, err = page.Reload()
	require.NoError(t, err)
	select {
	case <-firstFrame:
	case <-time.After(30 * time.Second):
		t.Fatal("should have received at least one frame")
	}

	err = screencast.Stop()
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Greater(t, len(frames), 0, "should have received at least one frame")
	require.Greater(t, len(frames[0]), 0, "frame data should not be empty")
}
