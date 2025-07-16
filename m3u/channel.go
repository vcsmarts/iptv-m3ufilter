package m3u

import "strings"

const DIRECTIVE_TRACK_INFO = "#EXTINF:"

type TVChannel struct {
	Info string
	URL  string
}

func LoadChannelsFromM3U(content string) []TVChannel {
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
