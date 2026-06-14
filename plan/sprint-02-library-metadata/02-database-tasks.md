# Sprint 02 — Veritabanı Görevleri (Medya Şeması)

> **Sprint hedefi:** Medya domain şeması (kütüphane, öğe, sezon/bölüm, dosya, metadata, kişiler).
>
> **İlgili:** [`../architecture.md`](../architecture.md) §4 · Sprint 01 [`../sprint-01-foundation/03-database-tasks.md`](../sprint-01-foundation/03-database-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (şema gözden geçirme)

## Bağımlılıklar
- Sprint 01 migrasyon altyapısı.

## Epikler ve Görevler

### Epik A: Çekirdek medya şeması
- [x] A1 — `library`, `media_item` (film/dizi/albüm), `season`, `episode`, `media_file` tabloları | Kabul: migrasyon + ilişki bütünlüğü.
- [x] A2 — Teknik metadata tabloları (kodek, akış/stream bilgileri) | Kabul: ffprobe çıktısı saklanır.
- [x] A3 — Görsel varlık (poster/backdrop) referans tablosu | Kabul: dosya/cache yolu tutulur.

### Epik B: Zenginleştirme şeması
- [x] B1 — Kişi (oyuncu/yönetmen), tür (genre), etiket tabloları + ilişkiler | Kabul: çoka-çok ilişkiler.
- [x] B2 — Harici sağlayıcı kimlikleri (TMDB/TVDB id eşleşmeleri) | Kabul: tekrar eşleştirmeyi destekler.
- [x] B3 — Çok dilli metin saklama (başlık/açıklama için locale bazlı; en-US + tr-TR) | Kabul: aynı öğe için dile göre metin sorgulanır. Bkz. [`../i18n-localization.md`](../i18n-localization.md).

### Epik C: İndeksleme ve performans
- [x] C1 — Arama/filtre için indeksler | Kabul: yaygın sorgular hızlı.
- [x] C2 — Tam metin arama temeli (SQLite FTS) | Kabul: başlık araması çalışır.

## Kabul Kriterleri (Sprint Çıktısı)
- Şema migrasyonlarla kurulur; tarama ve metadata verisi tutarlı yazılır.

## Riskler
- Şema sonraki sprintleri etkiler → ilişkiler dikkatli tasarlanmalı.

## Kapsam Dışı
- Kullanıcı/izleme durumu tabloları — Sprint 03.
