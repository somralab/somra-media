# Library

## Scanning

Somra watches library roots and runs scheduled full/incremental scans. File names are parsed for title, year, season, and episode.

- **Full scan:** re-walks all paths (use after bulk imports).
- **Incremental scan:** detects new/changed files only.

Scan progress is streamed to the UI. Large libraries may take several minutes on first scan.

## Metadata

Matched titles show posters, descriptions, and cast from configured providers (TMDB when API key is set). Unmatched items can be rematched manually from the detail page.

## File layout tips

```
movies/Movie Name (2020)/Movie Name (2020).mkv
tv/Show Name/Season 01/Show Name S01E01.mkv
```

See `testdata/library/README.md` for parser examples.

## Permissions

- **library:read** — browse and play
- **library:write** — create libraries, trigger scans, rematch

Admins have all permissions by default.
