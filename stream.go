package playwright

import (
	"bufio"
	"encoding/base64"
	"os"
	"path/filepath"
)

type streamImpl struct {
	channelOwner
}

func (s *streamImpl) SaveAs(path string) error {
	err := os.MkdirAll(filepath.Dir(path), 0o777)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for {
		binary, err := s.channel.Send("read", map[string]any{"size": 1024 * 1024})
		if err != nil {
			return err
		}
		bytes, err := base64.StdEncoding.DecodeString(binary.(string))
		if err != nil {
			return err
		}
		if len(bytes) == 0 {
			break
		}
		_, err = writer.Write(bytes)
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (s *streamImpl) ReadAll() ([]byte, error) {
	var data []byte
	for {
		binary, err := s.channel.Send("read", map[string]any{"size": 1024 * 1024})
		if err != nil {
			return nil, err
		}
		bytes, err := base64.StdEncoding.DecodeString(binary.(string))
		if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			break
		}
		data = append(data, bytes...)
	}
	return data, nil
}

func newStream(parent *channelOwner, objectType string, guid string, initializer map[string]any) *streamImpl {
	stream := &streamImpl{}
	stream.createChannelOwner(stream, parent, objectType, guid, initializer)
	return stream
}
