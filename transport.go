package playwright

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/go-jose/go-jose/v3/json"
)

type transport interface {
	Send(msg map[string]any) error
	Poll() (*message, error)
	Close() error
}

type pipeTransport struct {
	writer    io.WriteCloser
	bufReader *bufio.Reader
	closed    chan struct{}
	onClose   func() error
	process   *os.Process
}

func (t *pipeTransport) Poll() (*message, error) {
	if t.isClosed() {
		return nil, fmt.Errorf("transport closed")
	}

	var length uint32
	err := binary.Read(t.bufReader, binary.LittleEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("could not read protocol padding: %w", err)
	}

	data := make([]byte, length)
	_, err = io.ReadFull(t.bufReader, data)
	if err != nil {
		return nil, fmt.Errorf("could not read protocol data: %w", err)
	}

	msg := &message{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("could not decode json: %w", err)
	}
	if os.Getenv("DEBUGP") != "" {
		fmt.Fprintf(os.Stdout, "\x1b[33mRECV>\x1b[0m\n%s\n", data)
	}
	return msg, nil
}

type message struct {
	ID     int            `json:"id"`
	GUID   string         `json:"guid"`
	Method string         `json:"method,omitempty"`
	Params map[string]any `json:"params,omitempty"`
	Result map[string]any `json:"result,omitempty"`
	Error  *struct {
		Error Error `json:"error"`
	} `json:"error,omitempty"`
}

func (t *pipeTransport) Send(msg map[string]any) error {
	if t.isClosed() {
		return fmt.Errorf("transport closed")
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("pipeTransport: could not marshal json: %w", err)
	}
	if os.Getenv("DEBUGP") != "" {
		fmt.Fprintf(os.Stdout, "\x1b[32mSEND>\x1b[0m\n%s\n", msgBytes)
	}

	lengthPadding := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthPadding, uint32(len(msgBytes)))
	if _, err = t.writer.Write(append(lengthPadding, msgBytes...)); err != nil {
		return err
	}
	return nil
}

func (t *pipeTransport) Close() error {
	select {
	case <-t.closed:
		return nil
	default:
		return t.onClose()
	}
}

func (t *pipeTransport) isClosed() bool {
	select {
	case <-t.closed:
		return true
	default:
		return false
	}
}

func newPipeTransport(driver *PlaywrightDriver, stderr io.Writer) (transport, error) {
	t := &pipeTransport{
		closed: make(chan struct{}, 1),
	}

	cmd := driver.Command("run-driver")
	cmd.Stderr = stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not create stdout pipe: %w", err)
	}
	t.writer = stdin
	t.bufReader = bufio.NewReader(stdout)

	t.onClose = func() error {
		select {
		case <-t.closed:
		default:
			close(t.closed)
		}
		if err := t.writer.Close(); err != nil {
			return err
		}
		// playwright-cli will exit when its stdin is closed
		if err := cmd.Wait(); err != nil {
			return err
		}
		return nil
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start driver: %w", err)
	}

	t.process = cmd.Process

	return t, nil
}
