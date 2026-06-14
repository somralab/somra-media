# GPU hardware acceleration setup

Somra uses ffmpeg hardware encoders (Intel QSV, NVIDIA NVENC, VAAPI, AMD AMF) when
available. Software transcode (`libx264`) remains the fallback — playback continues
even when GPU paths fail.

## Quick start

1. Use the base compose file for CPU-only transcode:

   ```bash
   docker compose -f deploy/docker-compose.yml up --build
   ```

2. Add a GPU overlay that matches your hardware (see below).

3. In **Settings → Playback → Hardware acceleration**, leave mode on **Auto** (recommended).

## Intel / AMD — VAAPI and QSV (`/dev/dri`)

**Host requirements**

- Linux with `/dev/dri/renderD128` (or similar) present
- Intel: `intel-media-driver` (iHD) or legacy `intel-media-driver-legacy` (i965)
- AMD: `mesa-va-drivers` + amdgpu kernel module
- Docker user must access the render node (`video` or `render` group)

**Compose overlay**

```bash
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.vaapi.yml \
  up --build
```

**Environment overrides**

| Variable | Default | Purpose |
|----------|---------|---------|
| `LIBVA_DRIVER_NAME` | `iHD` | VAAPI driver (`iHD` Intel Gen8+, `i965` older Intel, `radeonsi` AMD) |
| `SOMRA_VAAPI_DEVICE` | `/dev/dri/renderD128` | Explicit render node |

**Verify**

```bash
docker exec somra ls -l /dev/dri
docker exec somra ffmpeg -hide_banner -encoders 2>/dev/null | grep -E 'h264_(qsv|vaapi)'
```

Somra prioritizes **Intel QSV** when `h264_qsv` is available, then VAAPI.

## NVIDIA — NVENC

**Host requirements**

- Proprietary NVIDIA driver
- [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html)
- `docker run --rm --gpus all nvidia/cuda:12.0.0-base-ubuntu22.04 nvidia-smi` succeeds

**Compose overlay**

```bash
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.nvidia.yml \
  up --build
```

**Environment overrides**

| Variable | Default | Purpose |
|----------|---------|---------|
| `NVIDIA_VISIBLE_DEVICES` | `all` | GPU selection (`0`, `1`, `all`) |
| `NVIDIA_GPU_COUNT` | `1` | Reserved GPU count for compose deploy stanza |

**Verify**

```bash
docker exec somra nvidia-smi
docker exec somra ffmpeg -hide_banner -encoders 2>/dev/null | grep h264_nvenc
```

## Multi-architecture images

Somra images are built for `linux/amd64` and `linux/arm64`. Hardware acceleration
availability depends on the **host** GPU and drivers, not the image architecture:

| Platform | Typical HW support |
|----------|------------------|
| amd64 + Intel iGPU | QSV / VAAPI via `/dev/dri` overlay |
| amd64 + NVIDIA | NVENC via NVIDIA toolkit overlay |
| arm64 (e.g. Apple Silicon VM, ARM SBC) | VAAPI when `/dev/dri` exposed; NVENC rare |
| macOS Docker Desktop | GPU passthrough limited — use software transcode |

The runtime image ships ffmpeg with VAAPI/QSV libraries (`intel-media-driver`, `libva`,
`mesa-va-drivers`). NVIDIA encode still requires the host driver + toolkit passthrough.

## Settings reference

| Setting | Values | Default |
|---------|--------|---------|
| Hardware mode | `auto`, `off`, `force` | `auto` |
| Preferred accelerator | `auto`, `qsv`, `nvenc`, `vaapi`, `amf` | `auto` |
| Max HW transcodes | 1–4 | 2 (when GPU detected) |
| Max concurrent transcodes | 1–8 | CPU-based smart default |

- **Auto**: use HW when detected and within session limits; fall back to CPU on error.
- **Off**: always `libx264`.
- **Force**: prefer HW; still falls back to CPU if the HW path fails (playback must not break).

## Troubleshooting

| Symptom | Likely cause | Action |
|---------|--------------|--------|
| No accelerators in onboarding | `/dev/dri` not passed or driver missing | Apply VAAPI overlay; check `ls /dev/dri` on host |
| NVENC missing in container | Toolkit not configured | Install nvidia-container-toolkit; use NVIDIA overlay |
| HW starts then falls back | Session limit or driver error | Check logs for `hw_fallback`; reduce concurrent HW sessions |
| Permission denied on `/dev/dri` | Group mismatch | Add host `video`/`render` GID via `group_add` in compose |

Structured logs include `hw_accelerator`, `hw_fallback`, and `hw_error` fields for diagnosis.
