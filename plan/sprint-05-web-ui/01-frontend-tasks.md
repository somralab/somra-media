# Sprint 05 — Frontend Görevleri (Kütüphane Gezinme & Keşif)

> **Sprint hedefi:** Zengin kütüphane gezinme, detay sayfaları, arama/filtre ve ana sayfa keşfi.
> Sprint 04 oynatıcısıyla birleşince M3 (kullanılabilir alfa) tamamlanır.
>
> **İlgili:** Sprint 02 (metadata), Sprint 04 (oynatıcı) · [`03-design-tasks.md`](./03-design-tasks.md) · [`02-backend-api-tasks.md`](./02-backend-api-tasks.md)

## Sorumlu Rol(ler)
- Frontend (birincil), UX/UI (tasarım)

## Bağımlılıklar
- Sprint 02 metadata + Sprint 03 kullanıcı + Sprint 04 oynatma API'leri.

## Epikler ve Görevler

### Epik A: Ana sayfa & keşif
- [ ] A1 — Ana sayfa: "izlemeye devam et", "yeni eklenenler", öneri rafları | Kabul: kullanıcı bazlı raflar.
- [ ] A2 — Kütüphane görünümü (grid/list, poster, lazy load, sanal liste) | Kabul: büyük kütüphane akıcı.

### Epik B: Detay sayfaları
- [ ] B1 — Film/dizi/sezon/bölüm detay sayfaları (metadata, oyuncular, görseller) | Kabul: zengin görünüm.
- [ ] B2 — "Oynat" / "devam et" / favori / izleme listesi aksiyonları | Kabul: backend'e bağlı.

### Epik C: Arama & filtre
- [ ] C1 — Hızlı arama (FTS) + sonuç önizleme | Kabul: anlık sonuç.
- [ ] C2 — Filtre/sıralama (tür, yıl, izlenme durumu) | Kabul: kombine filtreler.

### Epik D: Sorumluluk & durum
- [ ] D1 — Yükleme/boş/hata durumları + iskelet (skeleton) | Kabul: tutarlı UX.
- [ ] D2 — Ebeveyn kontrolüne göre içerik filtreleme (UI) | Kabul: çocuk profili kısıtlı görür.
- [ ] D3 — Tüm metinlerin i18n anahtarlarıyla (en-US + tr-TR) gelmesi ve metin uzunluğu/taşma dayanıklılığı | Kabul: hardcoded metin yok, TR↔EN'de düzen bozulmaz. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §5.
- [ ] D4 — Dört temanın (Cinematic varsayılan/Aurora/Noir/Minimal) tüm ekranlara uygulanması + hızlı tema değiştirici (menü/ayar) | Kabul: tema anında değişir, tüm ekranlarda tutarlı, seçim kalıcı (oturumsuz `localStorage`, oturumlu profil). Bkz. [`04-frontend-tasks.md`](../sprint-01-foundation/04-frontend-tasks.md) C1b ve [`03-design-tasks.md`](./03-design-tasks.md) Epik D.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı kütüphaneyi gezip arar, detay görür ve içeriği oynatır (uçtan uca alfa akışı).

## Riskler
- Büyük kütüphane performansı → sanal liste + sayfalama şart.

## Kapsam Dışı
- Kurulum sihirbazı — Sprint 06.
