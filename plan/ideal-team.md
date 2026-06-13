# Somra — İdeal Ekip Yapısı (Ideal Team)

> Bu doküman, "sıfırdan kendi motorumuz + tam özellik paritesi" hedefi için **gerçekçi ideal
> ekibi** tanımlar. [`project-brief.md`](./project-brief.md) kararına göre plan ideal ekip
> varsayımıyla hazırlanmıştır; ekip buna göre büyütülecektir.

İlgili: [`roadmap.md`](./roadmap.md) (kişi/iş dağılımı) · [`definition-of-done.md`](./definition-of-done.md)

---

## 1. Özet

| Ölçek | Toplam | Not |
|---|---|---|
| **İdeal ekip** | **~8–9 kişi** | Tam hızda paralel geliştirme. Bu planın varsayımı. |
| **Minimum çekirdek** | **3–4 kişi** | Mümkün ama takvim ~2–3x uzar; paralel iş azalır. |

Disiplinler her iki ölçekte de aynıdır; fark, paralel kapasite ve takvimdir.

## 2. Roller ve Sorumluluklar

### 2.1 Tech Lead / Yazılım Mimarı (Go) — 1 kişi
- Sistem mimarisi ve modül sınırları (bkz. [`architecture.md`](./architecture.md)).
- API sözleşmeleri, kod standardı ve [`definition-of-done.md`](./definition-of-done.md)'in sahibi.
- Teknik risk yönetimi, kütüphane/teknoloji kararları, kod inceleme son merci.
- Sprint planlama ve teknik backlog önceliklendirme (PM ile birlikte).
- **Çıktı sahibi:** Sprint 01 mimari görevleri.

### 2.2 Backend Mühendisi (Go) — 3 kişi
- Kütüphane tarama, metadata pipeline, kullanıcı/RBAC, istek yönetimi, otomasyon, indexer.
- Veri erişim katmanı, iş kuralları, arka plan işleri (job scheduler).
- Birim ve entegrasyon testleri.
- **Çıktı sahibi:** Sprint 02, 03, 06, 08, 09 backend görevlerinin çoğu.

### 2.3 Medya / Streaming Uzmanı (ffmpeg, kodek) — 1 kişi
- Transcode pipeline, HLS/DASH paketleme, adaptif bitrate, altyazı/ses kanalı işleme.
- Donanım hızlandırma (QSV/NVENC/VAAPI/AMF) tespiti ve seçimi.
- Oynatma uyumluluk matrisi ve kalite profilleri.
- **Çıktı sahibi:** Sprint 04, 07 medya görevleri.

### 2.4 Frontend Mühendisi (React) — 2 kişi
- React SPA (Vite), tasarım sistemi uygulaması, durum yönetimi, API entegrasyonu.
- Web video oynatıcı (hls.js/dash.js), kütüphane gezinme, arama, kurulum sihirbazı.
- Erişilebilirlik ve performans (lazy load, sanal liste).
- **Çıktı sahibi:** Sprint 05 ve diğer sprintlerin frontend görevleri.

### 2.5 DevOps / Platform Mühendisi — 1 kişi
- Tek Docker imajı + `docker compose` dağıtımı, çoklu mimari (amd64/arm64) build.
- CI/CD hattı, sürüm otomasyonu, imaj yayını (registry).
- Donanım cihaz erişimi (GPU passthrough), gözlemlenebilirlik (log/metrik).
- **Çıktı sahibi:** Sprint 01, 07, 10 devops görevleri.

### 2.6 QA / Test Otomasyon Mühendisi — 1 kişi
- Test stratejisi, e2e otomasyon, regresyon paketi, kabul testleri.
- Her sprintte DoD doğrulaması ve hata takibi.
- **Çıktı sahibi:** Her sprintteki QA görevleri.

### 2.7 UX/UI Tasarımcı — 0.5 kişi (yarı zamanlı)
- Tasarım sistemi, akışlar, kurulum sihirbazı UX'i, görsel kimlik (somra markası).
- **Çıktı sahibi:** Sprint 05 tasarım görevleri.

### 2.8 Ürün Yöneticisi / PM — 0.5 kişi (yarı zamanlı)
- Backlog, sprint yönetimi, kapsam koruma (anti-drift), sürüm planı.
- [`project-brief.md`](./project-brief.md) yönetişim kurallarının uygulanması.

## 3. Toplam Kadro Tablosu

| Rol | İdeal | Minimum çekirdek |
|---|---|---|
| Tech Lead / Mimar (Go) | 1 | 1 |
| Backend (Go) | 3 | 1 |
| Medya/Streaming | 1 | (Tech Lead/Backend paylaşır) |
| Frontend (React) | 2 | 1 |
| DevOps/Platform | 1 | (Tech Lead paylaşır) |
| QA | 1 | (geliştiriciler paylaşır) |
| UX/UI | 0.5 | (dış kaynak) |
| PM | 0.5 | (Tech Lead paylaşır) |
| **Toplam** | **~9 kişi** | **3–4 kişi** |

## 4. Beceri Matrisi (Skill Matrix)

| Beceri | Birincil rol | İkincil |
|---|---|---|
| Go servis mimarisi | Tech Lead | Backend |
| SQLite / veri modelleme | Backend | Tech Lead |
| ffmpeg / kodek / transcode | Medya Uzmanı | Tech Lead |
| Donanım hızlandırma (GPU) | Medya Uzmanı | DevOps |
| React / TypeScript | Frontend | — |
| Video oynatıcı (HLS/DASH) | Frontend | Medya Uzmanı |
| Docker / CI-CD | DevOps | Tech Lead |
| Test otomasyon | QA | tüm geliştiriciler |
| Güvenlik / RBAC | Backend | Tech Lead |
| UX / tasarım | UX/UI | Frontend |

## 5. Önerilen İşe Alım/Devreye Alma Sırası

1. **Tech Lead (Go)** — mimari ve standartların temeli (Sprint 01 öncesi).
2. **Backend #1 + DevOps** — iskelet, CI/CD, veri katmanı.
3. **Frontend #1** — tasarım sistemi ve API entegrasyon temeli.
4. **Medya/Streaming Uzmanı** — Sprint 04'ten önce.
5. **Backend #2–#3, Frontend #2, QA** — paralel kapasite arttıkça.
6. **UX/UI ve PM** — Sprint 01'den itibaren yarı zamanlı.

## 6. Çalışma Düzeni

- **Sprint kadansı:** 2 hafta (varsayılan). Katı deadline yok; kapsam koruma esastır.
- **Tören:** Sprint planlama, günlük kısa senkron, sprint demo (çalışan artımlı sürüm), retrospektif.
- **Kod inceleme:** Her PR en az 1 onay; mimariyi etkileyen değişiklikte Tech Lead onayı zorunlu.
- **Tek doğruluk kaynağı:** Kapsam ve karar tartışmalarında [`project-brief.md`](./project-brief.md) esastır.
