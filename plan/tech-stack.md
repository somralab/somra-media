# Somra — Teknoloji Yığını (Tech Stack)

> Onaylı teknoloji kararları ve gerekçeleri. Yeni bağımlılık eklenmesi Tech Lead onayına ve
> bu dokümanın güncellenmesine tabidir (anti-drift).

İlgili: [`architecture.md`](./architecture.md) · [`project-brief.md`](./project-brief.md) · [`definition-of-done.md`](./definition-of-done.md)

---

## 1. Backend

| Konu | Seçim | Gerekçe |
|---|---|---|
| Dil | **Go (1.x güncel)** | Tek binary, düşük bellek, yüksek eşzamanlılık, kolay dağıtım. |
| HTTP router | **`go-chi/chi`** | `net/http` uyumlu, hafif, idiomatic middleware; minimum bağımlılık. |
| Veritabanı | **SQLite (WAL)** | Gömülü, sıfır konfig, tek dosya. |
| SQLite sürücüsü | **`modernc.org/sqlite` (saf Go, CGO'suz)** | CGO gerektirmez → kolay çapraz derleme ve amd64/arm64 multi-arch build. |
| Migrasyon | **`pressly/goose`** (gömülü `embed.FS` migrasyonları) | Sürümlü şema, Go + SQL migrasyon, açılışta otomatik uygulama. |
| Medya analizi | **ffprobe** | Teknik metadata çıkarımı. |
| Transcode | **ffmpeg** | Endüstri standardı; imaja paketlenir. |
| İş zamanlama | **Kendi hafif scheduler'ımız + `robfig/cron/v3`** | Cron ifadeleri için cron kütüphanesi, durum/eşzamanlılık kontrolü kendi katmanımızda. |
| Kimlik/oturum | **JWT erişim token'ı (kısa ömürlü) + iptal edilebilir sunucu tarafı refresh token (DB)** | Stateless API + token iptali dengesi. |
| API sözleşmesi | **OpenAPI 3.1, design-first (elle yazılan spec)** | Tek doğruluk kaynağı; frontend TypeScript tipleri buradan üretilir. |
| i18n (backend) | **`nicksnyder/go-i18n/v2` + `golang.org/x/text`** | Locale-aware mesaj kataloğu, çoğul, dil eşleştirme. Bkz. [`i18n-localization.md`](./i18n-localization.md). |
| Test | Go `testing` + `testify` | Birim + entegrasyon. |

## 2. Frontend

| Konu | Seçim | Gerekçe |
|---|---|---|
| Çatı | **React + TypeScript** | Yaygın ekosistem, tip güvenliği. |
| Build | **Vite** | Hızlı dev/HMR, SPA çıktısı. |
| Durum yönetimi | **TanStack Query (sunucu durumu) + Zustand (UI/global durum)** | Sunucu durumu ve UI durumu net ayrımı, hafif. |
| Video oynatıcı | **hls.js** (+ Safari'de native HLS) | Tarayıcıda adaptif streaming. |
| Paketleme formatı | **CMAF (fMP4)** | Tek segment setinden HLS (birincil); DASH manifesti aynı segmentlerden opsiyonel. |
| Stil/tasarım sistemi | **Tailwind CSS + Radix UI primitifleri** | Kendi tasarım sistemimizi üzerine kurarız; erişilebilir, modern UI. |
| Tema sistemi | **Dinamik, kullanıcı-seçimli çoklu tema** (token tabanlı) | Özgün tema setleri: **Cinematic (varsayılan)**, Aurora, Noir, Minimal (marka taklidi yapılmaz). Tema kullanıcı bazında hatırlanır. Sprint 01/03/05. |
| i18n | **`i18next` + `react-i18next`** | Kaynak en-US, çeviri tr-TR; `Intl` ile tarih/sayı l10n. Bkz. [`i18n-localization.md`](./i18n-localization.md). |

## 3. Veri & Dağıtım

| Konu | Seçim |
|---|---|
| Birincil veri | SQLite |
| Dosya sistemi | Kullanıcı volume'ları (medya), önbellek/transcode dizini |
| Paketleme | Tek **Docker** imajı + `docker compose` örneği |
| Mimariler | amd64 + arm64 (multi-arch build) |
| Donanım erişimi | GPU passthrough (QSV/NVENC/VAAPI/AMF) — Sprint 07 |

## 4. CI/CD & Kalite

| Konu | Seçim |
|---|---|
| CI | Git tabanlı pipeline (lint + test + build + imaj) |
| Lint | Go: `golangci-lint`; Frontend: ESLint + Prettier |
| Sürüm | Anlamlı sürümleme (SemVer) + her sprintte artımlı milestone |
| İmaj yayını | **GitHub Container Registry (GHCR)** (birincil) + opsiyonel Docker Hub mirror |
| Katkı uzlaşısı | **DCO (Developer Certificate of Origin)** — CLA yok |

## 5. Harici Servisler / Sağlayıcılar

- **Metadata:** TMDB, TVDB, MusicBrainz, fanart.tv, OMDB (anahtar/oran sınırı yönetimi Sprint 02).
- **Altyazı:** Açık altyazı sağlayıcıları (Sprint 06).
- **Bildirim:** Webhook, Discord, e-posta (Sprint 08).

## 6. Bağımlılık Politikası

1. Minimum bağımlılık ilkesi: standart kütüphane öncelikli.
2. Her yeni bağımlılık: lisans uyumu (AGPL ile uyumlu olmalı), bakım durumu, güvenlik kontrolü.
3. Lisansı AGPL-3.0 ile uyumsuz bağımlılık **kullanılamaz** (bkz. [`project-brief.md`](./project-brief.md) §5).

## 7. Kapatılan Kararlar (Karar Verildi)

> Plan başlangıcında açık bırakılan tüm teknoloji kararları kapatılmıştır. Sprint 01 görevleri
> artık "karar ver" değil, "kararı uygula/doğrula" odaklıdır. Değişiklik yalnızca Tech Lead
> onayı + bu doküman güncellemesiyle yapılır (anti-drift).

| Karar | Sonuç |
|---|---|
| HTTP router | `go-chi/chi` |
| SQLite sürücüsü | `modernc.org/sqlite` (saf Go, CGO'suz) |
| Migrasyon aracı | `pressly/goose` (gömülü migrasyonlar) |
| Scheduler | Kendi hafif scheduler + `robfig/cron/v3` |
| Oturum stratejisi | JWT (kısa ömürlü) + iptal edilebilir refresh token (DB) |
| API sözleşmesi | OpenAPI 3.1 design-first → FE tip üretimi |
| Frontend durum yönetimi | TanStack Query + Zustand |
| Video oynatıcı / paketleme | hls.js + CMAF (HLS birincil, DASH opsiyonel) |
| Stil/tasarım sistemi | Tailwind CSS + Radix UI |
| Frontend i18n | `i18next` + `react-i18next` |
| Backend i18n | `nicksnyder/go-i18n/v2` + `x/text` |
| Çeviri platformu | **Weblate** (self-host, OSS, git entegre) |
| İmaj registry | GHCR (birincil) |
| Lisans / katkı | AGPL-3.0 + DCO |

Detaylar: oturum/sözleşme/eklenti için [`architecture.md`](./architecture.md) §8; i18n için [`i18n-localization.md`](./i18n-localization.md); lisans için [`project-brief.md`](./project-brief.md) §5.
