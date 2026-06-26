//go:build ignore
// +build ignore

package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func worker(id int, jobs chan Job, results chan<- Job, browser playwright.Browser) {
	for job := range jobs {
		fmt.Printf("starting (try: %d): %s\n", job.Try, job.URL)
		if job.Try >= 3 {
			job.Success = false
			job.err = fmt.Errorf("Stopped with domain %s (%w)", job.URL, job.err)
			results <- job
			continue
		}
		jobCtx, cancel := context.WithTimeout(context.Background(), time.Second*12)
		internalJobError := make(chan error, 1)
		go func() {
			internalJobError <- processJob(browser, job, jobCtx)
			cancel()
		}()
		select {
		case <-jobCtx.Done():
			job.err = fmt.Errorf("timeout (try: %d)", job.Try+1)
			job.Success = false
			job.Try++
			jobs <- job
		case err := <-internalJobError:
			if err != nil {
				job.err = err
				job.Success = false
				job.Try++
				jobs <- job
				cancel()
			} else {
				job.Success = true
				job.err = nil
				results <- job
			}
		}
	}
}

func processJob(browser playwright.Browser, job Job, ctx context.Context) error {
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"),
	})
	if err != nil {
		return fmt.Errorf("could not create context: %w", err)
	}
	defer context.Close()
	go func() {
		<-ctx.Done()
		context.Close()
	}()

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	_, err = page.Goto("http://"+job.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("could not goto: %s: %v", job.URL, err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get cwd %w", err)
	}
	_, err = page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(filepath.Join(cwd, "out", strings.Replace(job.URL, ".", "-", -1)+".png")),
	})
	if err != nil {
		return fmt.Errorf("could not screenshot: %w", err)
	}
	return nil
}

type Job struct {
	URL     string
	Try     int
	err     error
	Success bool
}

func main() {
	log.Println("Downloading Alexa top domains")
	topDomains, err := getAlexaTopDomains()
	assertErrorToNilf("could not get alexa top domains: %w", err)
	log.Println("Downloaded Alexa top domains successfully")
	cwd, err := os.Getwd()
	if err != nil {
		assertErrorToNilf("could not get cwd %w", err)
	}
	if err := os.Mkdir(filepath.Join(cwd, "out"), 0o777); err != nil && !os.IsExist(err) {
		assertErrorToNilf("could not create output directory %w", err)
	}

	pw, err := playwright.Run()
	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	assertErrorToNilf("could not launch Chromium: %w", err)

	numberOfJobs := int(math.Min(30, float64(len(topDomains))))

	jobs := make(chan Job, numberOfJobs)
	results := make(chan Job, numberOfJobs)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results, browser)
	}

	for _, url := range topDomains[:numberOfJobs] {
		jobs <- Job{
			URL: url,
		}
	}

	for a := 0; a < numberOfJobs; a++ {
		job := <-results
		if job.Success {
			fmt.Println("success:", job.URL)
		} else {
			fmt.Println("error:", job.URL, job.err)
		}
	}

	close(jobs)
	close(results)

	assertErrorToNilf("could not close browser: %w", browser.Close())
	assertErrorToNilf("could not stop Playwright: %w", pw.Stop())
}

func getAlexaTopDomains() ([]string, error) {
	resp, err := http.Get("http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip")
	if err != nil {
		return nil, fmt.Errorf("could not get: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %w", err)
	}
	defer resp.Body.Close()
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("could not create zip reader: %w", err)
	}
	alexaFile, err := zipReader.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("could not read alexa file: %w", err)
	}
	defer alexaFile.Close()
	reader := csv.NewReader(alexaFile)
	out := make([]string, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("could not read csv: %w", err)
		}
		out = append(out, record[1])
	}
}
