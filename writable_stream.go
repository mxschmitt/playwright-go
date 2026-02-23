package playwright

import (
	"encoding/base64"
	"io"
	"os"
)

type writableStream struct {
	channelOwner
}

func (s *writableStream) Copy(file string) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		buf := make([]byte, defaultCopyBufSize)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		_, err = s.channel.Send("write", map[string]any{
			"binary": base64.StdEncoding.EncodeToString(buf[:n]),
		})
		if err != nil {
			return err
		}
	}
	_, err = s.channel.Send("close")
	return err
}

func newWritableStream(parent *channelOwner, objectType string, guid string, initializer map[string]any) *writableStream {
	stream := &writableStream{}
	stream.createChannelOwner(stream, parent, objectType, guid, initializer)
	return stream
}
