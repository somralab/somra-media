# Sprint 09 — QA Görevleri

> **Sprint hedefi:** Eklenti mimarisi, indexer entegrasyonu ve uçtan uca otomasyonu doğrulamak.
> M5 (tam parite fazına giriş) kontrolü.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) · [`../roadmap.md`](../roadmap.md) (M5) · [`03-download-client-tasks.md`](./03-download-client-tasks.md)

## Sorumlu Rol(ler)
- QA (birincil)

## Bağımlılıklar
- Bu sprint tüm çıktıları.

## Epikler ve Görevler

### Epik A: Eklenti & izolasyon
- [ ] A1 — Eklentisiz çekirdek tam fonksiyon testi | Kabul: eklenti yokken sistem çalışır.
- [ ] A2 — Eklenti ekle/yapılandır/devre dışı testi | Kabul: yaşam döngüsü doğru.

### Epik B: Uçtan uca otomasyon
- [ ] B1 — İstek → kapma → indirme → import → kütüphane e2e | Kabul: akış tamamlanır.
- [ ] B2 — Kalite profili seçim doğruluğu | Kabul: doğru sürüm seçilir.
- [ ] B3 — Hata/yarıda kalma kurtarma testi | Kabul: sistem tutarlı kalır.

### Epik C: M5 kabul
- [ ] C1 — M5 kontrol listesi | Kabul: tam parite fazı kriterleri sağlanır.

## Kabul Kriterleri (Sprint Çıktısı)
- Otomasyon ve izolasyon test kapsamında; uçtan uca akış güvenilir; M5 kriterleri karşılanır.

## Riskler
- Harici istemci/indexer bağımlılığı → mock + kontrollü gerçek test ortamı.

## Kapsam Dışı
- Üretim güvenlik denetimi — Sprint 10.
