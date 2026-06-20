# Sprint 07 — Media & Streaming Tasks (Hardware Acceleration)

> **Sprint goal:** Hardware-accelerated transcode (QSV/NVENC/VAAPI/AMF) and automatic
> accelerator selection. M4 (beta candidate) core.
>
> **Related:** [`../project-brief.md`](../project-brief.md) (HW: software first, then HW) · Sprint 04 (CPU transcode pipeline) · [`02-devops-tasks.md`](./02-devops-tasks.md)

## Responsible Role(s)
- Media/Streaming Specialist (primary), DevOps (device access)

## Dependencies
- Sprint 04 (transcode pipeline abstraction), Sprint 06 (hardware detection).

## Epics and Tasks

### Epic A: Accelerator detection
- [x] A1 — Existing GPU/encoder detection (Intel QSV, NVIDIA NVENC/NVDEC, AMD/VAAPI/AMF) | Acceptance: available accelerators are listed.
- [x] A2 — Capability/supported codec matrix (HW decode/encode) | Acceptance: accurate capability report.

### Epic B: HW transcode pipeline
- [x] B1 — ffmpeg HW acceleration parameter generation (per platform) | Acceptance: HW transcode works.
- [x] B2 — HW decode + (if needed) HW encode full chain | Acceptance: CPU load drops significantly.
- [x] B3 — HW→SW fallback | Acceptance: falls back to CPU if HW fails; playback is not interrupted.

### Epic C: Automatic selection
- [x] C1 — Optimal path selection engine based on hardware + media | Acceptance: most efficient path is selected automatically.
- [x] C2 — Concurrent HW session limit (based on hardware constraints) | Acceptance: limit is not exceeded.

## Acceptance Criteria (Sprint Output)
- HW transcode works with at least one accelerator (priority: Intel QSV); automatic selection + fallback active.

## Risks
- **High technical risk.** Hardware/driver/container combinations are fragile → strong fallback and testing required.

## Out of Scope
- Guarantee for all GPU models — priority platforms are targeted; remainder is best-effort.
