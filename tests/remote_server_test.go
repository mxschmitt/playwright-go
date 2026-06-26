package playwright_test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mxschmitt/playwright-go"
)

type remoteServer struct {
	url string
	cmd *exec.Cmd
}

func newRemoteServer() (*remoteServer, error) {
	driver, err := playwright.NewDriver(&playwright.RunOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not start Playwright: %v", err)
	}
	cmd := driver.Command("run-server")
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not get stdout pipe: %v", err)
	}
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start server: %v", err)
	}
	scanner := bufio.NewReader(stdout)
	line, err := scanner.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read url: %v", err)
	}
	line = strings.TrimSpace(line)
	// Remove "Listening on " prefix
	const prefix = "Listening on "
	if !strings.HasPrefix(line, prefix) {
		return nil, fmt.Errorf("unexpected output format: %s", line)
	}
	url := strings.TrimPrefix(line, prefix)
	return &remoteServer{
		url: url,
		cmd: cmd,
	}, nil
}

func (s *remoteServer) Close() {
	_ = s.cmd.Process.Kill()
	_ = s.cmd.Wait()
}
