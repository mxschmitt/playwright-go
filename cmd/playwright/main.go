package main

import (
	"log"
	"os"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	driver, err := playwright.NewDriver(&playwright.RunOptions{})
	if err != nil {
		log.Fatalf("could not start driver: %v", err)
	}
	if err = driver.DownloadDriver(); err != nil {
		log.Fatalf("could not download driver: %v", err)
	}
	cmd := driver.Command(os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("could not run driver: %v", err)
	}
	os.Exit(cmd.ProcessState.ExitCode())
}
