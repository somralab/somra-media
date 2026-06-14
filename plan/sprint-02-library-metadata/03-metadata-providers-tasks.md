# Sprint 02 — Metadata Sağlayıcı Görevleri

> **Sprint hedefi:** Harici metadata sağlayıcılarından (TMDB/TVDB/MusicBrainz/fanart) zengin
> bilgi çekme, eşleştirme ve görsel indirme.
>
> **İlgili:** [`../tech-stack.md`](../tech-stack.md) §5 · [`../architecture.md`](../architecture.md) §3 · [`01-backend-tasks.md`](./01-backend-tasks.md) · [`../i18n-localization.md`](../i18n-localization.md)

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- [`01-backend-tasks.md`](./01-backend-tasks.md) (ön ayrıştırma çıktısı), [`02-database-tasks.md`](./02-database-tasks.md) (şema).

## Epikler ve Görevler

### Epik A: Sağlayıcı soyutlaması
- [x] A1 — Ortak `MetadataProvider` arayüzü (arama, detay, görseller) | Kabul: sağlayıcılar takılabilir.
- [x] A2 — API anahtarı yönetimi + oran sınırı (rate limit) + önbellek | Kabul: sınır aşımı engellenir, sonuçlar cache'lenir.

### Epik B: Sağlayıcı entegrasyonları
- [x] B1 — TMDB (film + dizi) | Kabul: doğru eşleşme oranı, test edilir.
- [x] B2 — TVDB (dizi) ve MusicBrainz (müzik) temel entegrasyonu | Kabul: temel alanlar çekilir.
- [x] B3 — fanart.tv / görsel sağlayıcı (poster/backdrop/logo) | Kabul: görseller indirilip cache'lenir.

### Epik C-dil: Çok dilli metadata
- [x] CL1 — Sağlayıcı sorgularında dil parametresi (kullanıcı/sistem locale'i: en-US/tr-TR) | Kabul: açıklama/başlık tercih edilen dilde çekilir, eksikse en-US'e düşülür. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §2.
- [x] CL2 — Çok dilli metadata saklama/önbellek stratejisi | Kabul: aynı öğe için TR+EN metin tutulabilir.

### Epik C: Eşleştirme (matching)
- [x] C1 — Ön ayrıştırma + sağlayıcı sonuçlarını eşleştirme algoritması (skorlama) | Kabul: yaygın durumlar doğru eşleşir.
- [x] C2 — Manuel düzeltme/yeniden eşleştirme API'si | Kabul: yanlış eşleşme düzeltilebilir.
- [x] C3 — Periyodik metadata yenileme işi (scheduler) | Kabul: güncellemeler alınır.

## Kabul Kriterleri (Sprint Çıktısı)
- Taranan öğeler zengin metadata + görsellerle eşleşir; manuel düzeltme mümkün.

## Riskler
- Sağlayıcı oran sınırları ve eşleşme doğruluğu → cache + skorlama önemli.

## Kapsam Dışı
- Altyazı indirme otomasyonu — Sprint 06.
