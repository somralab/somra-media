package library

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ProbeResult holds technical metadata extracted via ffprobe.
type ProbeResult struct {
	DurationMs    int64
	Container     string
	VideoCodec    string
	VideoWidth    int
	VideoHeight   int
	AudioCodec    string
	AudioChannels int
	SubtitleCount int
	RawJSON       string
}

type ffprobeOutput struct {
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Channels  int    `json:"channels"`
	} `json:"streams"`
}

// Prober runs ffprobe against media files.
type Prober struct {
	Binary string
}

// NewProber returns a prober using the given ffprobe binary path.
func NewProber(binary string) *Prober {
	if binary == "" {
		binary = "ffprobe"
	}
	return &Prober{Binary: binary}
}

// Probe extracts technical metadata from path.
func (p *Prober) Probe(ctx context.Context, path string) (ProbeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.Binary,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return ProbeResult{}, fmt.Errorf("ffprobe %q: %w", path, err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return ProbeResult{}, fmt.Errorf("ffprobe parse %q: %w", path, err)
	}

	result := ProbeResult{RawJSON: string(out)}
	if parsed.Format.Duration != "" {
		if secs, err := strconv.ParseFloat(parsed.Format.Duration, 64); err == nil {
			result.DurationMs = int64(secs * 1000)
		}
	}
	result.Container = parsed.Format.FormatName

	for _, s := range parsed.Streams {
		switch strings.ToLower(s.CodecType) {
		case "video":
			if result.VideoCodec == "" {
				result.VideoCodec = s.CodecName
				result.VideoWidth = s.Width
				result.VideoHeight = s.Height
			}
		case "audio":
			if result.AudioCodec == "" {
				result.AudioCodec = s.CodecName
				result.AudioChannels = s.Channels
			}
		case "subtitle":
			result.SubtitleCount++
		}
	}
	return result, nil
}
