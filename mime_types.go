package playwright

import (
	"fmt"
	"strings"
)

// mimeTypesByExtension mirrors the common subset of upstream's
// getMimeTypeForPath table (packages/isomorphic/mimeType.ts), using the exact
// MIME strings Playwright sends. We intentionally avoid Go's stdlib
// mime.TypeByExtension because it is platform-dependent and appends a charset.
var mimeTypesByExtension = map[string]string{
	"html":  "text/html",
	"htm":   "text/html",
	"css":   "text/css",
	"js":    "application/javascript",
	"mjs":   "application/javascript",
	"json":  "application/json",
	"xml":   "application/xml",
	"txt":   "text/plain",
	"csv":   "text/csv",
	"md":    "text/markdown",
	"ts":    "application/typescript",
	"jsx":   "text/jsx",
	"svg":   "image/svg+xml",
	"png":   "image/png",
	"jpg":   "image/jpeg",
	"jpeg":  "image/jpeg",
	"gif":   "image/gif",
	"webp":  "image/webp",
	"bmp":   "image/bmp",
	"ico":   "image/x-icon",
	"pdf":   "application/pdf",
	"zip":   "application/zip",
	"wasm":  "application/wasm",
	"woff":  "font/woff",
	"woff2": "font/woff2",
	"ttf":   "font/ttf",
	"otf":   "font/otf",
	"mp4":   "video/mp4",
	"webm":  "video/webm",
	"ogv":   "video/ogg",
	"mov":   "video/quicktime",
	"mpeg":  "video/mpeg",
	"mp3":   "audio/mpeg",
	"wav":   "audio/wav",
	"ogg":   "audio/ogg",
	"oga":   "audio/ogg",
	"m4a":   "audio/mp4",
}

// getMimeTypeForPath returns the MIME type for a file path based on its
// extension, mirroring upstream getMimeTypeForPath. Returns an empty string
// when the extension is unknown so callers can fall back to
// "application/octet-stream".
func getMimeTypeForPath(path string) string {
	dotIndex := strings.LastIndex(path, ".")
	if dotIndex == -1 {
		return ""
	}
	// Case-sensitive lookup, matching upstream getMimeTypeForPath (its table is
	// all-lowercase, so a ".PNG" extension yields no match → octet-stream, and
	// determineScreenshotType errors on it rather than silently inferring png).
	extension := path[dotIndex+1:]
	return mimeTypesByExtension[extension]
}

// determineScreenshotType infers the screenshot type from the output path's
// extension when no explicit type is set, mirroring upstream
// determineScreenshotType. It errors on an unsupported extension.
func determineScreenshotType(path *string, typ *ScreenshotType) (*ScreenshotType, error) {
	if typ != nil {
		return typ, nil
	}
	if path == nil {
		return nil, nil
	}
	mimeType := getMimeTypeForPath(*path)
	switch mimeType {
	case "image/png":
		return ScreenshotTypePng, nil
	case "image/jpeg":
		return ScreenshotTypeJpeg, nil
	default:
		return nil, fmt.Errorf("path: unsupported mime type %q", mimeType)
	}
}
