package m3u

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const DIRECTIVE_TRACK_INFO = "#EXTINF:"

type TVChannel struct {
	Info string
	URL  string
}

type IPTVFilter struct {
	Client  http.Client
	Timeout time.Duration
}

func (f *IPTVFilter) testStream(url string) bool {

	ctx, cancel := context.WithTimeout(context.Background(), f.Timeout)
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

	return res.StatusCode == http.StatusOK
}

func (f *IPTVFilter) FilterWorkingStreams(tvChannels []TVChannel) []TVChannel {
	var workingChannels []TVChannel

	for i, ch := range tvChannels {
		fmt.Printf("Testing channel %d of %d\n", i+1, len(tvChannels))
		if f.testStream(ch.URL) {
			workingChannels = append(workingChannels, ch)
		}
	}
	return workingChannels
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
