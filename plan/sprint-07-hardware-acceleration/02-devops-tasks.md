# Sprint 07 — DevOps Görevleri (GPU Passthrough & Paketleme)

> **Sprint hedefi:** Docker içinde GPU/cihaz erişimi, hızlandırıcı sürücü/runtime entegrasyonu
> ve imaj uyumluluğu.
>
> **İlgili:** Sprint 01 [`../sprint-01-foundation/05-devops-tasks.md`](../sprint-01-foundation/05-devops-tasks.md) · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md)

## Sorumlu Rol(ler)
- DevOps/Platform (birincil), Medya Uzmanı (doğrulama)

## Bağımlılıklar
- Sprint 01 Docker imajı, [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md).

## Epikler ve Görevler

### Epik A: Cihaz erişimi
- [x] A1 — Intel/AMD VAAPI için `/dev/dri` passthrough + compose örneği | Kabul: konteyner GPU'ya erişir.
- [x] A2 — NVIDIA için container toolkit/runtime entegrasyonu + örnek | Kabul: NVENC kullanılabilir.

### Epik B: İmaj uyumluluğu
- [x] B1 — ffmpeg'in HW hızlandırma destekli derlenmesi/paketlenmesi | Kabul: imajda HW kodlayıcılar mevcut.
- [x] B2 — Multi-arch + HW uyumluluk doğrulaması | Kabul: hedef platformlarda çalışır.

### Epik C: Dokümantasyon
- [x] C1 — Kullanıcı için GPU kurulum rehberi (compose örnekleri) | Kabul: net adımlar.

## Kabul Kriterleri (Sprint Çıktısı)
- Docker'da en az bir GPU yolu (öncelik: Intel `/dev/dri`) sorunsuz çalışır ve belgelidir.

## Riskler
- Host sürücü bağımlılığı → net önkoşullar ve fallback dokümante edilmeli.

## Kapsam Dışı
- macOS VideoToolbox (Docker'da kısıtlı) — best-effort, garanti yok.
