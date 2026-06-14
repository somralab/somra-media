# Library naming test fixtures (Sprint 02 QA)

These paths document expected parser output for `internal/library.ParseFileName`.
They are not executed directly; see `internal/library/parser_test.go` and
`testdata/library/naming/manifest.json`.

## Movies

- `movies/Inception (2010)/Inception (2010).mkv` → title "Inception", year 2010
- `movies/The.Matrix.1999.1080p.mkv` → title "The Matrix", year 1999

## TV

- `tv/Breaking Bad/Season 01/Breaking Bad S01E02.mkv` → S01E02
- `tv/Show Name/Show Name 2x05.mkv` → 2x05 pattern

## Music

- `music/Artist - Album/01 Track.flac` → title from filename (basic)

## Corrupt / unsupported

- `edge/corrupt/not-a-video.txt` — skipped by scan (non-media extension)
- `edge/corrupt/truncated.mkv` — ffprobe failure; scan continues
