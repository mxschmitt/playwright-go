package playwright_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/h2non/filetype"
	"github.com/playwright-community/playwright-go"
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

func TestScreencastStartStop(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)

	screencast, err := page.Screencast()
	require.NoError(t, err)

	var frames [][]byte
	err = screencast.Start(playwright.ScreencastStartOptions{
		OnFrame: func(frame playwright.OnFrame) {
			frames = append(frames, frame.Data)
		},
	})
	require.NoError(t, err)

	// Trigger some activity to generate frames
	_, err = page.Reload()
	require.NoError(t, err)
	//nolint:staticcheck
	page.WaitForTimeout(500)

	err = screencast.Stop()
	require.NoError(t, err)

	require.Greater(t, len(frames), 0, "should have received at least one frame")
	require.Greater(t, len(frames[0]), 0, "frame data should not be empty")
}
