# Sprint 09 — İndirme & Otomasyon Görevleri (*arr İşlevi)

> **Sprint hedefi:** İndirme istemcisi entegrasyonu + kalite profilleri + otomatik kapma (grab)
> ve içe aktarma (import) (Sonarr/Radarr işlevi). İstek yönetimiyle uçtan uca otomasyon.
>
> **İlgili:** [`02-indexer-integration-tasks.md`](./02-indexer-integration-tasks.md) · Sprint 08 (istek handoff) · Sprint 02 (kütüphane/import)

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- [`01-plugin-architecture-tasks.md`](./01-plugin-architecture-tasks.md), [`02-indexer-integration-tasks.md`](./02-indexer-integration-tasks.md), Sprint 08 istek akışı.

## Epikler ve Görevler

### Epik A: İndirme istemcisi adaptörleri
- [ ] A1 — Torrent + Usenet indirme istemcisi adaptör arayüzü (ekle, durum, tamamlandı) | Kabul: yaygın istemciler eklenebilir.
- [ ] A2 — İndirme durumu izleme (scheduler) | Kabul: ilerleme/tamamlanma takip edilir.

### Epik B: Kalite profilleri & karar
- [ ] B1 — Kalite profili tanımı (çözünürlük/kodek/boyut tercihleri) | Kabul: profil bazlı seçim.
- [ ] B2 — Indexer sonuçlarını profile göre skorlama + otomatik kapma | Kabul: en iyi sürüm seçilir.

### Epik C: İçe aktarma & izleme listeleri
- [ ] C1 — Tamamlanan indirmeyi kütüphaneye import (yeniden adlandırma/taşıma + tarama tetikleme) | Kabul: medya kütüphanede belirir.
- [ ] C2 — İzleme listesi/monitör (dizi bölümleri otomatik takip) | Kabul: yeni bölüm otomatik aranır.
- [ ] C3 — İstek → onay → otomatik edinme uçtan uca akışı (Sprint 08 ile bağlama) | Kabul: onaylı istek otomatik tamamlanır.

## Kabul Kriterleri (Sprint Çıktısı)
- Onaylı bir istek; indexer arama → kalite seçimi → indirme → import → kütüphane akışını otomatik tamamlar.

## Riskler
- Karmaşık uçtan uca akış + yasal hassasiyet → eklenti izolasyonu ve sağlam hata yönetimi.

## Kapsam Dışı
- Tam Lidarr/Readarr (müzik/kitap) paritesi — bu plan kapsamı dışında (bkz. brief §7).
