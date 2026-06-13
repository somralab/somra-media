# Sprint 06 — Backend Görevleri (Akıllı Varsayılanlar & Ayarlar)

> **Sprint hedefi:** "Minimum konfig / maksimum optimizasyon" felsefesini hayata geçiren
> sistem tespiti, akıllı varsayılan üretimi ve merkezi ayar yönetimi.
>
> **İlgili:** [`../project-brief.md`](../project-brief.md) (başarı kriterleri) · [`../architecture.md`](../architecture.md) (Ayarlar & Onboarding) · [`02-frontend-wizard-tasks.md`](./02-frontend-wizard-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil), Medya Uzmanı (transcode profili önerileri)

## Bağımlılıklar
- Sprint 01 (ayar katmanı), Sprint 02 (kütüphane), Sprint 04 (transcode profilleri).

## Epikler ve Görevler

### Epik A: Sistem tespiti
- [ ] A1 — Donanım tespiti (CPU, bellek, mevcut GPU varlığı) | Kabul: sistem profili çıkarılır.
- [ ] A2 — Depolama/dizin tespiti ve doğrulama (medya/cache yolları) | Kabul: yazma/okuma izni doğrulanır.

### Epik B: Akıllı varsayılanlar
- [ ] B1 — Donanıma göre transcode profili/eşzamanlılık önerisi (CPU bazlı; GPU Sprint 07'de genişler) | Kabul: makul varsayılan üretilir.
- [ ] B2 — Önerilen kütüphane tarama/ yenileme zamanlaması | Kabul: varsayılan job programı.

### Epik C: Merkezi ayar yönetimi
- [ ] C1 — Ayar şeması + API (kategori bazlı, doğrulamalı) | Kabul: ayarlar tek yerden yönetilir.
- [ ] C2 — Kurulum durumu (onboarding tamamlandı mı) state machine | Kabul: ilk kurulum akışını yönetir.
- [ ] C3 — Sistem varsayılan dili ayarı (tr-TR/en-US) | Kabul: kullanıcı tercihi yoksa bu kullanılır; dil pazarlığı önceliklerine uyar. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §3.

## Kabul Kriterleri (Sprint Çıktısı)
- Sistem kendini tespit eder, akıllı varsayılan üretir; ayarlar merkezi API ile yönetilir.

## Riskler
- Yanlış varsayılan kötü deneyim → muhafazakâr ama optimize varsayılanlar + override.

## Kapsam Dışı
- GPU bazlı optimizasyon detayı — Sprint 07.
