package streaming

import (
	"fmt"
	"strings"
)

// WriteMasterPlaylist builds an HLS master manifest for the given variants.
func WriteMasterPlaylist(variants []VariantInfo) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:7\n")
	for _, v := range variants {
		fmt.Fprintf(&b, "#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,NAME=\"%s\"\n",
			v.Bandwidth, v.Width, v.Height, v.Name)
		b.WriteString(v.MediaPlaylist + "\n")
	}
	return b.String()
}

// VariantInfo describes one HLS rendition.
type VariantInfo struct {
	Name          string
	Bandwidth     int64
	Width         int
	Height        int
	MediaPlaylist string
}

// WriteMediaPlaylist builds a VOD media playlist referencing fMP4 segments.
func WriteMediaPlaylist(targetDuration int, initURI string, segments []SegmentRef) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:7\n")
	b.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	b.WriteString("#EXT-X-TARGETDURATION:" + fmt.Sprintf("%d", targetDuration) + "\n")
	b.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")
	fmt.Fprintf(&b, "#EXT-X-MAP:URI=\"%s\"\n", initURI)
	for _, seg := range segments {
		fmt.Fprintf(&b, "#EXTINF:%.3f,\n", seg.DurationSec)
		b.WriteString(seg.URI + "\n")
	}
	b.WriteString("#EXT-X-ENDLIST\n")
	return b.String()
}

// SegmentRef is one HLS segment entry.
type SegmentRef struct {
	URI         string
	DurationSec float64
}

// InitSegmentName is the standard CMAF init segment filename.
const InitSegmentName = "init.mp4"

// SegmentName formats a CMAF segment filename.
func SegmentName(index int) string {
	return fmt.Sprintf("seg_%05d.m4s", index)
}
