# Sprint 01 — DevOps Görevleri

> **Sprint hedefi:** Tek Docker imajı iskeleti, CI/CD hattı ve geliştirme ortamı standardı.
>
> **İlgili:** [`../tech-stack.md`](../tech-stack.md) §3–§4 · [`../definition-of-done.md`](../definition-of-done.md) §5

## Sorumlu Rol(ler)
- DevOps/Platform (birincil), Tech Lead (gözetim)

## Bağımlılıklar
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) ve [`02-backend-tasks.md`](./02-backend-tasks.md) (derlenebilir servis).

## Epikler ve Görevler

### Epik A: Docker imajı
- [x] A1 — Çok aşamalı (multi-stage) Dockerfile: Go binary + frontend statik + ffmpeg | Kabul: imaj boyutu makul, servis ayağa kalkar.
- [x] A2 — `docker-compose.yml` örneği (volume'lar: config, medya, transcode/cache) | Kabul: tek komutla çalışır.
- [x] A3 — Çoklu mimari build (amd64 + arm64) | Kabul: her iki mimaride imaj üretilir.

### Epik B: CI/CD hattı
- [x] B1 — CI pipeline: lint → i18n-check → unit-test → integration-test → coverage-gate → build → image-build | Kabul: DoD §5 kapıları uygulanır.
- [x] B1b — `i18n-check` adımı: eksik/kullanılmayan anahtar + en-US/tr-TR tamlık kontrolü | Kabul: eksik çeviri PR'ı kırar. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §6.
- [x] B1c — `coverage-gate` adımı: Go + frontend coverage ölçümü, rapor üretimi ve eşik kapısı (çekirdek ≥%80, kritik modüller ≥%90, frontend bileşen ≥%70) | Kabul: eşik altında merge engellenir; rapor PR'a eklenir. Bkz. [`../definition-of-done.md`](../definition-of-done.md) §4.1.
- [x] B2 — Frontend lint/test/build entegrasyonu | Kabul: FE adımları yeşil.
- [x] B3 — Sürüm etiketleme + imaj yayını iskeleti (registry kararı) | Kabul: etiketli imaj yayınlanır.

### Epik C: Geliştirme ortamı
- [x] C1 — Yerel geliştirme akışı (hot reload backend + frontend dev server) | Kabul: README ile dokümante.
- [x] C2 — `Makefile`/görev koşucusu (build, test, lint, run) | Kabul: standart komutlar çalışır.

## Kabul Kriterleri (Sprint Çıktısı)
- `docker compose up` ile servis ayağa kalkar; health ucu yanıt verir.
- CI tüm aşamalarıyla yeşil; her PR bu kapılardan geçer.

## Riskler
- ffmpeg paketleme + multi-arch build karmaşıklığı → erken doğrulama.

## Kapsam Dışı
- GPU passthrough (donanım hızlandırma) — Sprint 07.
