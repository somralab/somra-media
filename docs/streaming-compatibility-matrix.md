# Streaming compatibility matrix (Sprint 04)

| Container | Video | Audio | Expected mode | Notes |
|-----------|-------|-------|---------------|-------|
| MP4 | H.264 | AAC | direct_play | Browser-safe |
| MKV | H.264 | AAC | direct_stream | Remux to CMAF/HLS |
| MP4 | HEVC | AAC | transcode | libx264 + AAC |
| MKV | H.264 | AC3 | transcode | Unsupported audio |

Run `go test ./internal/streaming/... -run TestDecisionEngine` for automated decision checks.
