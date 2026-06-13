# Sprint 02 — QA Görevleri

> **Sprint hedefi:** Tarama ve metadata pipeline'ının doğruluğunu ve dayanıklılığını doğrulamak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`01-backend-tasks.md`](./01-backend-tasks.md) · [`03-metadata-providers-tasks.md`](./03-metadata-providers-tasks.md)

## Sorumlu Rol(ler)
- QA (birincil), Backend (test verisi)

## Bağımlılıklar
- Sprint 01 test çatıları.

## Epikler ve Görevler

### Epik A: Tarama testleri
- [ ] A1 — Çeşitli adlandırma/klasör yapısı test seti | Kabul: ayrıştırma doğruluğu ölçülür.
- [ ] A2 — Bozuk/desteklenmeyen dosya dayanıklılık testi | Kabul: tarama çökmeden devam eder.
- [ ] A3 — Büyük kütüphane performans testi (temel) | Kabul: süre/kaynak raporlanır.

### Epik B: Metadata testleri
- [ ] B1 — Sağlayıcı eşleştirme doğruluğu (mock + gerçek örnek) | Kabul: kabul edilebilir eşleşme oranı.
- [ ] B2 — Oran sınırı/cache davranışı testi | Kabul: sınır aşımı yaşanmaz.

### Epik C: Regresyon
- [ ] C1 — Sprint 02 regresyon paketi | Kabul: CI'da koşar.

## Kabul Kriterleri (Sprint Çıktısı)
- Tarama/metadata akışları test kapsamında; kritik hata yok.

## Riskler
- Harici sağlayıcı testleri kırılgan → mock + sınırlı gerçek test dengesi.

## Kapsam Dışı
- Oynatma/streaming testleri — Sprint 04.
