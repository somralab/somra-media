# Sprint 07 — Medya & Streaming Görevleri (Donanım Hızlandırma)

> **Sprint hedefi:** Donanım hızlandırmalı transcode (QSV/NVENC/VAAPI/AMF) ve otomatik
> hızlandırıcı seçimi. M4 (beta adayı) çekirdeği.
>
> **İlgili:** [`../project-brief.md`](../project-brief.md) (HW: önce yazılım, sonra HW) · Sprint 04 (CPU transcode pipeline) · [`02-devops-tasks.md`](./02-devops-tasks.md)

## Sorumlu Rol(ler)
- Medya/Streaming Uzmanı (birincil), DevOps (cihaz erişimi)

## Bağımlılıklar
- Sprint 04 (transcode pipeline soyutlaması), Sprint 06 (donanım tespiti).

## Epikler ve Görevler

### Epik A: Hızlandırıcı tespiti
- [x] A1 — Mevcut GPU/encoder tespiti (Intel QSV, NVIDIA NVENC/NVDEC, AMD/VAAPI/AMF) | Kabul: kullanılabilir hızlandırıcılar listelenir.
- [x] A2 — Yetenek/desteklenen kodek matrisi (HW decode/encode) | Kabul: doğru yetenek raporu.

### Epik B: HW transcode pipeline
- [x] B1 — ffmpeg HW hızlandırma parametre üretimi (her platform için) | Kabul: HW transcode çalışır.
- [x] B2 — HW decode + (gerekiyorsa) HW encode tam zincir | Kabul: CPU yükü belirgin düşer.
- [x] B3 — HW→SW geri düşüş (fallback) | Kabul: HW başarısızsa CPU'ya düşer, oynatma kesilmez.

### Epik C: Otomatik seçim
- [x] C1 — Donanım + medyaya göre en uygun yol seçim motoru | Kabul: en verimli yol otomatik seçilir.
- [x] C2 — Eşzamanlı HW oturum limiti (donanım sınırına göre) | Kabul: limit aşılmaz.

## Kabul Kriterleri (Sprint Çıktısı)
- En az bir hızlandırıcı (öncelik: Intel QSV) ile HW transcode çalışır; otomatik seçim + fallback aktif.

## Riskler
- **Yüksek teknik risk.** Donanım/sürücü/konteyner kombinasyonları kırılgan → güçlü fallback ve test şart.

## Kapsam Dışı
- Tüm GPU modellerinin garantisi — öncelikli platformlar hedeflenir, kalanı best-effort.
