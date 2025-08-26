package main

import (
	"flag"
	"fmt"
	"log"
	"m3ufilter/m3u"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var (
		m3uFileUrl = flag.String("m3u-file", "", "URL with the m3u file.")
		outputDir  = flag.String("output", "./filtered", "Output directory for filtered M3U files.")
		timeout    = flag.Int("timeout", 5, "timeout in seconds for each http request. Must be between 1-5.")
		maxWorkers = flag.Int("max-workers", 2, "how many requests workers/requests can be done at same time.")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "IPTV Filter CLI Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Examples: \n")
		fmt.Fprintf(os.Stderr, "	%s -m3u-file https://example.com/playlist.m3u\n", os.Args[0])
	}

	flag.Parse()

	parsedURL, err := url.Parse(*m3uFileUrl)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		log.Printf("Invalid URL: %s\n", parsedURL)
		os.Exit(1)
	}
	if *timeout < 1 || *timeout > 5 {
		log.Printf("invalid timeout\n")
		os.Exit(1)
	}
	if *maxWorkers <= 0 {
		log.Printf("invalid amount of workers. It must be at least 1.")
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		os.Exit(1)
	}

	filter := m3u.IPTVFilter{
		Client:     http.Client{},
		Timeout:    time.Duration(*timeout) * time.Second,
		MaxWorkers: *maxWorkers,
	}

	m3uContent, err := filter.DownloadM3U(*m3uFileUrl)
	if err != nil {
		log.Printf("[--] ❌ Error downloading M3U: %v\n", *m3uFileUrl)
		os.Exit(1)
	}
	tvChannels := filter.LoadChannelsFromM3U(m3uContent)
	filteredTvChannels := filter.FilterWorkingStreams(tvChannels)
	outputPath := filepath.Join(*outputDir, "filtered.m3u")

	if err := filter.SaveFilteredM3U(filteredTvChannels, outputPath); err != nil {
		log.Printf("[%s] ❌ Error saving filtered M3U: %v", outputPath, err)
	} else {
		log.Printf("[%s] ✅ Saved %d working channels to %s", outputPath, len(filteredTvChannels), outputPath)
	}

	log.Printf("from %d to %d channels", len(tvChannels), len(filteredTvChannels))
}
