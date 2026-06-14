package streaming

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PackagerOptions configures HLS segment generation via ffmpeg.
type PackagerOptions struct {
	SourcePath          string
	OutputDir           string
	Mode                Mode
	StartPositionMs     int64
	AudioStreamIndex    *int
	SubtitleStreamIndex *int
	Tiers               []LadderTier
	SegmentSeconds      int
}

// BuildFFmpegArgs returns ffmpeg arguments for the given packaging mode.
func BuildFFmpegArgs(opts PackagerOptions) []string {
	seg := opts.SegmentSeconds
	if seg <= 0 {
		seg = 4
	}
	tier := LadderTier{Name: "source", Width: 0, Height: 0, VideoBitrate: 2_500_000, AudioBitrate: 128_000}
	if len(opts.Tiers) > 0 {
		tier = opts.Tiers[0]
	}

	args := []string{}
	if opts.StartPositionMs > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", float64(opts.StartPositionMs)/1000.0))
	}
	args = append(args, "-i", opts.SourcePath)

	if opts.AudioStreamIndex != nil {
		args = append(args, "-map", "0:a:"+fmt.Sprintf("%d", *opts.AudioStreamIndex))
	} else {
		args = append(args, "-map", "0:a:0?")
	}
	args = append(args, "-map", "0:v:0?")

	initPath := filepath.Join(opts.OutputDir, InitSegmentName)
	segPattern := filepath.Join(opts.OutputDir, "seg_%05d.m4s")

	switch opts.Mode {
	case ModeDirectStream:
		args = append(args,
			"-c", "copy",
			"-f", "hls",
			"-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", InitSegmentName,
			"-hls_time", fmt.Sprintf("%d", seg),
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", segPattern,
			filepath.Join(opts.OutputDir, "stream.m3u8"),
		)
	case ModeTranscode:
		args = append(args,
			"-c:v", "libx264", "-preset", "veryfast", "-profile:v", "baseline",
			"-pix_fmt", "yuv420p",
		)
		if tier.Width > 0 && tier.Height > 0 {
			args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", tier.Width, tier.Height))
		}
		args = append(args,
			"-b:v", fmt.Sprintf("%d", tier.VideoBitrate),
			"-c:a", "aac", "-b:a", fmt.Sprintf("%d", tier.AudioBitrate), "-ac", "2",
			"-f", "hls",
			"-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", InitSegmentName,
			"-hls_time", fmt.Sprintf("%d", seg),
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", segPattern,
			filepath.Join(opts.OutputDir, "stream.m3u8"),
		)
	default:
		_ = initPath
	}
	return args
}

// WriteDirectPlayManifest creates a minimal master playlist for progressive source.
func WriteDirectPlayManifest(outputDir string) error {
	media := WriteMediaPlaylist(4, "source", []SegmentRef{{URI: "source", DurationSec: 1.0}})
	return os.WriteFile(filepath.Join(outputDir, "master.m3u8"), []byte(media), 0o644)
}

// EnsureOutputDir creates the session cache directory.
func EnsureOutputDir(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("streaming mkdir %q: %w", dir, err)
	}
	return nil
}

// StartPackaging launches ffmpeg packaging when needed.
func StartPackaging(ctx context.Context, pm *ProcessManager, opts PackagerOptions) error {
	if opts.Mode == ModeDirectPlay {
		return WriteDirectPlayManifest(opts.OutputDir)
	}
	if err := EnsureOutputDir(opts.OutputDir); err != nil {
		return err
	}
	args := BuildFFmpegArgs(opts)
	if len(args) == 0 {
		return fmt.Errorf("streaming packager: no args for mode %q", opts.Mode)
	}
	_, err := pm.Start(ctx, filepath.Base(opts.OutputDir), args)
	return err
}

// WaitForManifest polls until master or stream playlist exists.
func WaitForManifest(dir string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, name := range []string{"master.m3u8", "stream.m3u8"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("streaming manifest timeout in %q", dir)
}
