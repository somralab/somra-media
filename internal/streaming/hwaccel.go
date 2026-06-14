package streaming

import (
	"fmt"
	"os"
	"strings"
)

// Accelerator identifies a hardware transcode backend.
type Accelerator string

const (
	AccelQSV   Accelerator = "qsv"
	AccelNVENC Accelerator = "nvenc"
	AccelVAAPI Accelerator = "vaapi"
	AccelAMF   Accelerator = "amf"
	AccelNone  Accelerator = ""
)

// HWMode controls hardware acceleration policy.
type HWMode string

const (
	HWModeAuto  HWMode = "auto"
	HWModeOff   HWMode = "off"
	HWModeForce HWMode = "force"
)

// TranscodePath describes the selected encode route.
type TranscodePath struct {
	UseHW         bool
	Accelerator   Accelerator
	VideoEncoder  string
	FallbackToSW  bool
	SelectionNote string
}

// HWRuntimeConfig holds live HW transcode limits and policy.
type HWRuntimeConfig struct {
	Mode             HWMode
	Preferred        Accelerator
	Available        []Accelerator
	MaxHWSessions    int
	MaxTotalSessions int
	VAAPIDevice      string
}

// DefaultHWRuntimeConfig returns conservative software-only defaults.
func DefaultHWRuntimeConfig() HWRuntimeConfig {
	return HWRuntimeConfig{
		Mode:             HWModeAuto,
		MaxHWSessions:    2,
		MaxTotalSessions: 2,
		VAAPIDevice:      vaapiDevice(),
	}
}

func vaapiDevice() string {
	if v := os.Getenv("SOMRA_VAAPI_DEVICE"); v != "" {
		return v
	}
	return "/dev/dri/renderD128"
}

// acceleratorPriority defines auto-selection order (Intel QSV first).
var acceleratorPriority = []Accelerator{AccelQSV, AccelNVENC, AccelVAAPI, AccelAMF}

// SelectTranscodePath picks HW or SW encode based on policy, availability, and media.
func SelectTranscodePath(cfg HWRuntimeConfig, probe MediaProbe, activeHW int) TranscodePath {
	if cfg.Mode == HWModeOff {
		return TranscodePath{UseHW: false, VideoEncoder: "libx264", SelectionNote: "hw_disabled"}
	}

	avail := availableSet(cfg.Available)
	preferred := cfg.Preferred
	if preferred == AccelNone || preferred == "auto" {
		preferred = pickBest(avail)
	}

	if preferred == AccelNone {
		if cfg.Mode == HWModeForce {
			return TranscodePath{UseHW: false, VideoEncoder: "libx264", FallbackToSW: true, SelectionNote: "hw_unavailable_force_fallback"}
		}
		return TranscodePath{UseHW: false, VideoEncoder: "libx264", SelectionNote: "no_hw_available"}
	}

	if cfg.MaxHWSessions > 0 && activeHW >= cfg.MaxHWSessions {
		return TranscodePath{UseHW: false, VideoEncoder: "libx264", SelectionNote: "hw_session_limit"}
	}

	if !avail[preferred] {
		if cfg.Mode == HWModeForce {
			return TranscodePath{UseHW: false, VideoEncoder: "libx264", FallbackToSW: true, SelectionNote: "preferred_hw_missing"}
		}
		preferred = pickBest(avail)
		if preferred == AccelNone {
			return TranscodePath{UseHW: false, VideoEncoder: "libx264", SelectionNote: "no_hw_available"}
		}
	}

	enc := hwEncoderFor(preferred)
	if enc == "" {
		return TranscodePath{UseHW: false, VideoEncoder: "libx264", SelectionNote: "encoder_unsupported"}
	}

	return TranscodePath{
		UseHW:         true,
		Accelerator:   preferred,
		VideoEncoder:  enc,
		SelectionNote: "hw_selected",
	}
}

func availableSet(ids []Accelerator) map[Accelerator]bool {
	m := make(map[Accelerator]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

func pickBest(avail map[Accelerator]bool) Accelerator {
	for _, id := range acceleratorPriority {
		if avail[id] {
			return id
		}
	}
	return AccelNone
}

func hwEncoderFor(acc Accelerator) string {
	switch acc {
	case AccelQSV:
		return "h264_qsv"
	case AccelNVENC:
		return "h264_nvenc"
	case AccelVAAPI:
		return "h264_vaapi"
	case AccelAMF:
		return "h264_amf"
	default:
		return ""
	}
}

// AppendHWVideoArgs adds platform-specific ffmpeg arguments for HW transcode.
func AppendHWVideoArgs(args []string, path TranscodePath, tier LadderTier, vaapiDevice string) []string {
	if !path.UseHW {
		return appendSWVideoArgs(args, tier)
	}
	if vaapiDevice == "" {
		vaapiDevice = vaapiDeviceDefault()
	}

	switch path.Accelerator {
	case AccelQSV:
		args = append(args,
			"-hwaccel", "qsv", "-hwaccel_output_format", "qsv",
			"-c:v", "h264_qsv", "-preset", "veryfast", "-global_quality", "23",
		)
	case AccelNVENC:
		args = append(args,
			"-hwaccel", "cuda", "-hwaccel_output_format", "cuda",
			"-c:v", "h264_nvenc", "-preset", "p4", "-tune", "hq",
		)
	case AccelVAAPI:
		scale := ""
		if tier.Width > 0 && tier.Height > 0 {
			scale = fmt.Sprintf("scale_vaapi=w=%d:h=%d", tier.Width, tier.Height)
		}
		vf := "format=nv12,hwupload"
		if scale != "" {
			vf = fmt.Sprintf("%s,%s", vf, scale)
		}
		args = append(args,
			"-hwaccel", "vaapi", "-hwaccel_device", vaapiDevice, "-hwaccel_output_format", "vaapi",
			"-vf", vf,
			"-c:v", "h264_vaapi", "-qp", "23",
		)
	case AccelAMF:
		args = append(args,
			"-c:v", "h264_amf", "-quality", "balanced",
		)
	default:
		return appendSWVideoArgs(args, tier)
	}

	if path.Accelerator != AccelVAAPI && tier.Width > 0 && tier.Height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", tier.Width, tier.Height))
	}
	args = append(args, "-b:v", fmt.Sprintf("%d", tier.VideoBitrate))
	return args
}

func appendSWVideoArgs(args []string, tier LadderTier) []string {
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-profile:v", "baseline",
		"-pix_fmt", "yuv420p",
	)
	if tier.Width > 0 && tier.Height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", tier.Width, tier.Height))
	}
	args = append(args, "-b:v", fmt.Sprintf("%d", tier.VideoBitrate))
	return args
}

func vaapiDeviceDefault() string {
	return vaapiDevice()
}

// AcceleratorsFromSettings converts settings accelerator IDs to streaming types.
func AcceleratorsFromSettings(ids []string) []Accelerator {
	out := make([]Accelerator, 0, len(ids))
	for _, id := range ids {
		switch strings.ToLower(id) {
		case string(AccelQSV), string(AccelNVENC), string(AccelVAAPI), string(AccelAMF):
			out = append(out, Accelerator(strings.ToLower(id)))
		}
	}
	return out
}
