package settings

import (
	"context"
	"strconv"

	"github.com/somralab/somra-media/internal/streaming"
)

// StreamingRuntimeConfig maps persisted settings to streaming HW runtime values.
type StreamingRuntimeConfig struct {
	MaxConcurrentTranscodes int
	MaxHWTranscodes         int
	HWMode                  HWMode
	HWAccelerator           AcceleratorID
	AvailableAccelerators   []AcceleratorID
}

// GetStreamingRuntimeConfig builds runtime streaming limits from DB + ffmpeg probe.
func (s *Service) GetStreamingRuntimeConfig(ctx context.Context, ffmpegBin string) (StreamingRuntimeConfig, error) {
	maxConcurrent, err := s.intSetting(ctx, KeyStreamingMaxConcurrent, 2)
	if err != nil {
		return StreamingRuntimeConfig{}, err
	}
	maxHW, err := s.intSetting(ctx, KeyStreamingMaxHW, 2)
	if err != nil {
		return StreamingRuntimeConfig{}, err
	}
	modeRaw, err := s.GetString(ctx, KeyStreamingHWMode, string(HWModeAuto))
	if err != nil {
		return StreamingRuntimeConfig{}, err
	}
	accelRaw, err := s.GetString(ctx, KeyStreamingHWAccelerator, "auto")
	if err != nil {
		return StreamingRuntimeConfig{}, err
	}

	profile := DetectSystemWithFFmpeg(nil, ffmpegBin)
	available := make([]AcceleratorID, 0, len(profile.Accelerators))
	for _, a := range profile.Accelerators {
		if a.Available {
			available = append(available, a.ID)
		}
	}

	mode := HWMode(modeRaw)
	if mode != HWModeAuto && mode != HWModeOff && mode != HWModeForce {
		mode = HWModeAuto
	}

	return StreamingRuntimeConfig{
		MaxConcurrentTranscodes: maxConcurrent,
		MaxHWTranscodes:         maxHW,
		HWMode:                  mode,
		HWAccelerator:           AcceleratorID(accelRaw),
		AvailableAccelerators:   available,
	}, nil
}

// ToStreamingHWConfig converts settings runtime config to streaming package config.
func (c StreamingRuntimeConfig) ToStreamingHWConfig() streaming.HWRuntimeConfig {
	avail := make([]streaming.Accelerator, 0, len(c.AvailableAccelerators))
	for _, id := range c.AvailableAccelerators {
		avail = append(avail, streaming.Accelerator(id))
	}
	preferred := streaming.Accelerator(c.HWAccelerator)
	if c.HWAccelerator == "" || c.HWAccelerator == "auto" {
		preferred = streaming.AccelNone
	}
	return streaming.HWRuntimeConfig{
		Mode:             streaming.HWMode(c.HWMode),
		Preferred:        preferred,
		Available:        avail,
		MaxHWSessions:    c.MaxHWTranscodes,
		MaxTotalSessions: c.MaxConcurrentTranscodes,
	}
}

func (s *Service) intSetting(ctx context.Context, key string, fallback int) (int, error) {
	raw, err := s.GetString(ctx, key, strconv.Itoa(fallback))
	if err != nil {
		return 0, err
	}
	n, convErr := strconv.Atoi(raw)
	if convErr != nil || n < 1 {
		return fallback, nil
	}
	return n, nil
}
