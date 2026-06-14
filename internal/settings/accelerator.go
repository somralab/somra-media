package settings

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// AcceleratorID identifies a hardware transcode backend.
type AcceleratorID string

const (
	AcceleratorQSV   AcceleratorID = "qsv"
	AcceleratorNVENC AcceleratorID = "nvenc"
	AcceleratorVAAPI AcceleratorID = "vaapi"
	AcceleratorAMF   AcceleratorID = "amf"
)

// HWMode controls hardware acceleration policy.
type HWMode string

const (
	HWModeAuto  HWMode = "auto"
	HWModeOff   HWMode = "off"
	HWModeForce HWMode = "force"
)

// AcceleratorInfo describes a detected ffmpeg HW backend.
type AcceleratorInfo struct {
	ID            AcceleratorID `json:"id"`
	Available     bool          `json:"available"`
	DevicePresent bool          `json:"devicePresent"`
	EncodeCodecs  []string      `json:"encodeCodecs"`
	DecodeCodecs  []string      `json:"decodeCodecs"`
}

// acceleratorSpec maps logical accelerators to ffmpeg encoder/decoder names.
var acceleratorSpecs = []struct {
	id       AcceleratorID
	encoders []string
	decoders []string
}{
	{id: AcceleratorQSV, encoders: []string{"h264_qsv", "hevc_qsv"}, decoders: []string{"h264_qsv", "hevc_qsv"}},
	{id: AcceleratorNVENC, encoders: []string{"h264_nvenc", "hevc_nvenc"}, decoders: []string{"h264_cuvid", "hevc_cuvid"}},
	{id: AcceleratorVAAPI, encoders: []string{"h264_vaapi", "hevc_vaapi"}, decoders: []string{"h264_vaapi", "hevc_vaapi"}},
	{id: AcceleratorAMF, encoders: []string{"h264_amf", "hevc_amf"}, decoders: []string{"h264_amf", "hevc_amf"}},
}

// acceleratorPriority defines auto-selection order (Intel QSV first).
var acceleratorPriority = []AcceleratorID{
	AcceleratorQSV,
	AcceleratorNVENC,
	AcceleratorVAAPI,
	AcceleratorAMF,
}

var ffmpegProbeFn = probeFFmpegOutput

// ProbeAccelerators inspects ffmpeg encoders/hwaccels and host devices.
func ProbeAccelerators(ffmpegBin string) []AcceleratorInfo {
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}
	encoders := parseCodecSet(ffmpegProbeFn(context.Background(), ffmpegBin, "-encoders"))
	decoders := parseCodecSet(ffmpegProbeFn(context.Background(), ffmpegBin, "-decoders"))
	hwaccels := parseHWAccels(ffmpegProbeFn(context.Background(), ffmpegBin, "-hwaccels"))

	out := make([]AcceleratorInfo, 0, len(acceleratorSpecs))
	for _, spec := range acceleratorSpecs {
		info := AcceleratorInfo{ID: spec.id, DevicePresent: devicePresentFor(spec.id)}
		for _, enc := range spec.encoders {
			if encoders[enc] {
				info.EncodeCodecs = append(info.EncodeCodecs, enc)
			}
		}
		for _, dec := range spec.decoders {
			if decoders[dec] || hwaccels[strings.TrimSuffix(dec, "_cuvid")] || hwaccels[string(spec.id)] {
				info.DecodeCodecs = append(info.DecodeCodecs, dec)
			}
		}
		info.Available = len(info.EncodeCodecs) > 0
		if spec.id == AcceleratorNVENC {
			info.DevicePresent = detectNVIDIADevice()
		}
		out = append(out, info)
	}
	return out
}

// RecommendAccelerator picks the best available accelerator for auto mode.
func RecommendAccelerator(accelerators []AcceleratorInfo) AcceleratorID {
	byID := make(map[AcceleratorID]AcceleratorInfo, len(accelerators))
	for _, a := range accelerators {
		byID[a.ID] = a
	}
	for _, id := range acceleratorPriority {
		if a, ok := byID[id]; ok && a.Available && a.DevicePresent {
			return id
		}
	}
	for _, id := range acceleratorPriority {
		if a, ok := byID[id]; ok && a.Available {
			return id
		}
	}
	return ""
}

func devicePresentFor(id AcceleratorID) bool {
	switch id {
	case AcceleratorQSV, AcceleratorVAAPI, AcceleratorAMF:
		return detectDRIDevice()
	case AcceleratorNVENC:
		return detectNVIDIADevice()
	default:
		return false
	}
}

func detectDRIDevice() bool {
	entries, err := os.ReadDir("/dev/dri")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "renderD") {
			return true
		}
	}
	return false
}

func detectNVIDIADevice() bool {
	if _, err := os.Stat("/dev/nvidia0"); err == nil {
		return true
	}
	if runtime.GOOS == "linux" {
		out, err := exec.CommandContext(context.Background(), "nvidia-smi", "-L").Output()
		return err == nil && len(strings.TrimSpace(string(out))) > 0
	}
	return false
}

func probeFFmpegOutput(ctx context.Context, bin string, flag string) string {
	cmd := exec.CommandContext(ctx, bin, "-hide_banner", flag)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return string(out)
}

func parseCodecSet(output string) map[string]bool {
	found := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "V") && !strings.HasPrefix(line, "A") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		found[fields[1]] = true
	}
	return found
}

func parseHWAccels(output string) map[string]bool {
	found := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Hardware") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		found[fields[0]] = true
	}
	return found
}
