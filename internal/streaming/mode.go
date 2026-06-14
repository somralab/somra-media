package streaming

// Mode describes how media is delivered to the client.
type Mode string

const (
	ModeDirectPlay   Mode = "direct_play"
	ModeDirectStream Mode = "direct_stream"
	ModeTranscode    Mode = "transcode"
)

// Decision captures the output of the playback decision engine.
type Decision struct {
	Mode   Mode
	Reason string
}
