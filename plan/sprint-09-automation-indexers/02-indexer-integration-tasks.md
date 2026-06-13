# Sprint 09 — Indexer Entegrasyon Görevleri (Prowlarr İşlevi)

> **Sprint hedefi:** Torrent + Usenet indexer entegrasyonu (Prowlarr işlevi): arama, yetenek
> sorgulama ve sonuç normalizasyonu — eklenti çerçevesi üzerinden.
>
> **İlgili:** [`01-plugin-architecture-tasks.md`](./01-plugin-architecture-tasks.md) · [`../project-brief.md`](../project-brief.md) §4 (içerik edinme: ileri sprint)

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- [`01-plugin-architecture-tasks.md`](./01-plugin-architecture-tasks.md) (eklenti sözleşmesi).

## Epikler ve Görevler

### Epik A: Indexer soyutlaması
- [ ] A1 — Torrent + Usenet için ortak arama/yetenek arayüzü | Kabul: tür-bağımsız sonuç modeli.
- [ ] A2 — Indexer tanımı/şema (Torznab/Newznab benzeri uyumluluk) | Kabul: yaygın protokoller desteklenir.

### Epik B: Arama & normalizasyon
- [ ] B1 — Çoklu indexer'da paralel arama + sonuç birleştirme | Kabul: tekilleştirilmiş sonuç.
- [ ] B2 — Sonuç ayrıştırma (kalite, çözünürlük, boyut, seeder) | Kabul: skorlamaya hazır alanlar.

### Epik C: Yönetim
- [ ] C1 — Indexer ekleme/test/devre dışı API'si | Kabul: bağlantı testi çalışır.

## Kabul Kriterleri (Sprint Çıktısı)
- Birden fazla indexer eklenip aranabilir; sonuçlar normalize ve test edilir.

## Riskler
- Protokol/indexer çeşitliliği + yasal hassasiyet → uyumluluk + izolasyon.

## Kapsam Dışı
- Kapma/indirme yürütme — [`03-download-client-tasks.md`](./03-download-client-tasks.md).
