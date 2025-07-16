package m3u

import (
	"reflect"
	"testing"
)

func TestLoadChannelsFromM3U(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []TVChannel
	}{
		{
			name: "valid m3u with single channel",
			content: `#EXTM3U
#EXTINF:-1 info1
https://example.com/index1.m3u8`,
			want: []TVChannel{{
				Info: "#EXTINF:-1 info1",
				URL:  "https://example.com/index1.m3u8",
			}},
		},
		{
			name: "valid m3u with multiple channels",
			content: `#EXTM3U
#EXTINF:-1,Channel 1
http://example.com/stream1.m3u8
#EXTINF:-1,Channel 2
http://example.com/stream2.m3u8
#EXTINF:-1,Channel 3
http://example.com/stream3.m3u8`,
			want: []TVChannel{
				{
					Info: "#EXTINF:-1,Channel 1",
					URL:  "http://example.com/stream1.m3u8",
				},
				{
					Info: "#EXTINF:-1,Channel 2",
					URL:  "http://example.com/stream2.m3u8",
				},
				{
					Info: "#EXTINF:-1,Channel 3",
					URL:  "http://example.com/stream3.m3u8",
				},
			},
		},
		{
			name: "extinf without url",
			content: `#EXTM3U
#EXTINF:-1,Channel 1`,
			want: nil,
		},
		{
			name: "extinf with empty url",
			content: `#EXTM3U
#EXTINF:-1,Channel 1
`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoadChannelsFromM3U(tt.content)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadChannelsFromM3U() = %v, want %v", got, tt.want)
			}
		})
	}

}
