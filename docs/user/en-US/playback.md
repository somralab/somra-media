# Playback

Somra plays media in the browser via HLS (CMAF). The server chooses **direct play** when the client supports the codecs, otherwise **transcode**.

## Starting playback

1. Open a title from **Libraries** or **Home**.
2. Click **Play**.
3. The player loads automatically; use keyboard shortcuts (space, arrows, F for fullscreen).

## Transcoding

- Software transcode uses ffmpeg; limit concurrent sessions in **Settings → Playback**.
- Hardware acceleration (QSV/NVENC/VAAPI) is auto-detected when enabled — see [gpu-setup.md](../gpu-setup.md).

## Subtitles

On the detail page, search and attach subtitles. Burned-in subs require transcode.

## Troubleshooting

- **Spinner never ends:** check `/api/v1/health` → `ffmpeg` and `transcode` checks.
- **Buffering:** lower quality or enable HW accel; verify cache disk space.
