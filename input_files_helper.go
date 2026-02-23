package playwright

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const fileSizeLimitInBytes = 50 * 1024 * 1024

var ErrInputFilesSizeExceeded = errors.New("Cannot set buffer larger than 50Mb, please write it to a file and pass its path instead.")

type inputFiles struct {
	Selector        *string             `json:"selector,omitempty"`
	Streams         []*channel          `json:"streams,omitempty"` // writableStream
	LocalPaths      []string            `json:"localPaths,omitempty"`
	Payloads        []map[string]string `json:"payloads,omitempty"`
	LocalDirectory  *string             `json:"localDirectory,omitempty"`
	DirectoryStream *channel            `json:"directoryStream,omitempty"`
}

type fileItem struct {
	LastModifiedMs *int64 `json:"lastModifiedMs,omitempty"`
	Name           string `json:"name"`
}

// convertInputFiles converts files to proper format for Playwright
//
//   - files should be one of: string, []string, InputFile, []InputFile,
//     string: local file path
func convertInputFiles(files any, context *browserContextImpl) (*inputFiles, error) {
	var (
		converted = &inputFiles{}
		paths     []string
	)
	switch items := files.(type) {
	case InputFile:
		if sizeOfInputFiles([]InputFile{items}) > fileSizeLimitInBytes {
			return nil, ErrInputFilesSizeExceeded
		}
		converted.Payloads = normalizeFilePayloads([]InputFile{items})
	case []InputFile:
		if sizeOfInputFiles(items) > fileSizeLimitInBytes {
			return nil, ErrInputFilesSizeExceeded
		}
		converted.Payloads = normalizeFilePayloads(items)
	case string: // local file path
		paths = []string{items}
	case []string:
		paths = items
	default:
		return nil, errors.New("files should be one of: string, []string, InputFile, []InputFile")
	}

	localPaths, localDir, err := resolvePathsAndDirectoryForInputFiles(paths)
	if err != nil {
		return nil, err
	}

	if !context.connection.isRemote {
		converted.LocalPaths = localPaths
		converted.LocalDirectory = localDir
		return converted, nil
	}

	// remote
	params := map[string]any{
		"items": []fileItem{},
	}
	allFiles := localPaths
	if localDir != nil {
		params["rootDirName"] = filepath.Base(*localDir)
		allFiles, err = listFiles(*localDir)
		if err != nil {
			return nil, err
		}
	}
	for _, file := range allFiles {
		lastModifiedMs, err := getFileLastModifiedMs(file)
		if err != nil {
			return nil, fmt.Errorf("failed to get last modified time of %s: %w", file, err)
		}
		filename := filepath.Base(file)
		if localDir != nil {
			var err error
			filename, err = filepath.Rel(*localDir, file)
			if err != nil {
				return nil, err
			}
		}
		params["items"] = append(params["items"].([]fileItem), fileItem{
			LastModifiedMs: &lastModifiedMs,
			Name:           filename,
		})
	}

	ret, err := context.connection.WrapAPICall(func() (any, error) {
		return context.channel.SendReturnAsDict("createTempFiles", params)
	}, true)
	if err != nil {
		return nil, err
	}
	result := ret.(map[string]any)

	streams := make([]*channel, 0)
	items := result["writableStreams"].([]any)
	for i := 0; i < len(allFiles); i++ {
		stream := fromChannel(items[i]).(*writableStream)
		if err := stream.Copy(allFiles[i]); err != nil {
			return nil, err
		}
		streams = append(streams, stream.channel)
	}

	if result["rootDir"] != nil {
		if ch, ok := result["rootDir"].(*channel); ok {
			converted.DirectoryStream = ch
		} else {
			converted.DirectoryStream = fromChannel(result["rootDir"]).(*writableStream).channel
		}
	} else {
		converted.Streams = streams
	}

	return converted, nil
}

func getFileLastModifiedMs(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if info.IsDir() {
		return 0, fmt.Errorf("%s is a directory", path)
	}
	return info.ModTime().UnixMilli(), nil
}

func sizeOfInputFiles(files []InputFile) int {
	size := 0
	for _, file := range files {
		size += len(file.Buffer)
	}
	return size
}

func normalizeFilePayloads(files []InputFile) []map[string]string {
	out := make([]map[string]string, 0)
	for _, file := range files {
		out = append(out, map[string]string{
			"name":     file.Name,
			"mimeType": file.MimeType,
			"buffer":   base64.StdEncoding.EncodeToString(file.Buffer),
		})
	}
	return out
}

func resolvePathsAndDirectoryForInputFiles(items []string) (localPaths []string, localDirectory *string, e error) {
	for _, item := range items {
		abspath, err := filepath.Abs(item)
		if err != nil {
			e = err
			return
		}
		// if the path is a directory
		if info, err := os.Stat(abspath); err == nil {
			if info.IsDir() {
				if localDirectory != nil {
					e = errors.New("Multiple directories are not supported")
					return
				}
				localDirectory = &abspath
			} else {
				if localPaths == nil {
					localPaths = []string{abspath}
				} else {
					localPaths = append(localPaths, abspath)
				}
			}
		} else {
			e = err
			return
		}
	}
	if localPaths != nil && localDirectory != nil {
		e = errors.New("File paths must be all files or a single directory")
	}
	return
}

func listFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
