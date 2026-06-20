# Sprint 07 — DevOps Tasks (GPU Passthrough & Packaging)

> **Sprint goal:** GPU/device access inside Docker, accelerator driver/runtime integration,
> and image compatibility.
>
> **Related:** Sprint 01 [`../sprint-01-foundation/05-devops-tasks.md`](../sprint-01-foundation/05-devops-tasks.md) · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md)

## Responsible Role(s)
- DevOps/Platform (primary), Media Specialist (validation)

## Dependencies
- Sprint 01 Docker image, [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md).

## Epics and Tasks

### Epic A: Device access
- [x] A1 — `/dev/dri` passthrough for Intel/AMD VAAPI + compose example | Acceptance: container accesses GPU.
- [x] A2 — NVIDIA container toolkit/runtime integration + example | Acceptance: NVENC is available.

### Epic B: Image compatibility
- [x] B1 — ffmpeg build/packaging with HW acceleration support | Acceptance: HW encoders present in image.
- [x] B2 — Multi-arch + HW compatibility validation | Acceptance: works on target platforms.

### Epic C: Documentation
- [x] C1 — GPU setup guide for users (compose examples) | Acceptance: clear steps.

## Acceptance Criteria (Sprint Output)
- At least one GPU path (priority: Intel `/dev/dri`) works smoothly in Docker and is documented.

## Risks
- Host driver dependency → clear prerequisites and fallback must be documented.

## Out of Scope
- macOS VideoToolbox (limited in Docker) — best-effort, no guarantee.
