package m3u

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const DIRECTIVE_TRACK_INFO = "#EXTINF:"
const JOBS_RESULTS_BUFFER_MULTIPLIER = 3

type TVChannel struct {
	Info string
	URL  string
}

type IPTVFilter struct {
	Client     http.Client
	Timeout    time.Duration
	MaxWorkers int
}

func (f *IPTVFilter) quickHeadCheck(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), f.Timeout/2)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "VLC/3.0.0 LibVLC/3.0.0")

	res, err := f.Client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	// Accept both OK and partial content responses
	return res.StatusCode == http.StatusOK || res.StatusCode == http.StatusPartialContent
}

func (f *IPTVFilter) verifyStreamData(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), f.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("[%.50s] Error occurred: %s ", url, err.Error())
		return false
	}
	req.Header.Set("User-Agent", "VLC/3.0.0 LibVLC/3.0.0")
	req.Header.Set("Range", "bytes=0-4095") // First 4KB

	res, err := f.Client.Do(req)
	if err != nil {
		log.Printf("[%.50s] Error occurred: %s ", url, err.Error())
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		return false
	}

	// Read a small amount to verify data flows
	buffer := make([]byte, 1024)
	_, err = res.Body.Read(buffer)
	return err == nil || err == io.EOF
}

func (f *IPTVFilter) testStream(url string) bool {
	if !f.quickHeadCheck(url) {
		return false
	}
	return f.verifyStreamData(url)
}

func (f *IPTVFilter) worker(jobChannel <-chan TVChannel, jobResultChannel chan<- TVChannel, wg *sync.WaitGroup) {
	defer wg.Done()
	for tvChannel := range jobChannel {
		if f.testStream(tvChannel.URL) {
			jobResultChannel <- tvChannel
		}
	}
}

func (f *IPTVFilter) FilterWorkingStreams(tvChannels []TVChannel) (workingChannels []TVChannel) {

	jobChannel := make(chan TVChannel, f.MaxWorkers)
	jobResultChannel := make(chan TVChannel, f.MaxWorkers*JOBS_RESULTS_BUFFER_MULTIPLIER)
	wg := sync.WaitGroup{}

	for i := 0; i < f.MaxWorkers; i++ {
		wg.Add(1)
		go f.worker(jobChannel, jobResultChannel, &wg)
	}

	// send jobs
	go func() {
		defer close(jobChannel)
		for _, tvChannel := range tvChannels {
			jobChannel <- tvChannel
		}
	}()

	// close results channel once all workers are done
	// results sent before close can still be read after the close.
	go func() {
		wg.Wait()
		close(jobResultChannel)
	}()

	// collect results
	for workingTvChannel := range jobResultChannel {
		workingChannels = append(workingChannels, workingTvChannel)
		log.Printf("[✓] %d/%d - Working: %.50s...",
			len(workingChannels), len(tvChannels), workingTvChannel.URL)
	}
	return workingChannels
}

func (f *IPTVFilter) DownloadM3U(url string) (string, error) {
	resp, err := f.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (f *IPTVFilter) LoadChannelsFromM3U(content string) []TVChannel {
	var tvChannels []TVChannel
	lines := strings.Split(content, "\n")

	for pos, line := range lines {
		if strings.HasPrefix(line, DIRECTIVE_TRACK_INFO) {
			if pos+1 < len(lines) {
				url := strings.TrimSpace(lines[pos+1])
				if strings.HasPrefix(url, "http") {
					tvChannels = append(tvChannels, TVChannel{
						Info: line,
						URL:  url,
					})
				}

			}
		}
	}
	return tvChannels
}

func (f *IPTVFilter) SaveFilteredM3U(channels []TVChannel, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	writer.WriteString("#EXTM3U\n")
	for _, channel := range channels {
		writer.WriteString(channel.Info + "\n")
		writer.WriteString(channel.URL + "\n")
	}

	return nil
}
