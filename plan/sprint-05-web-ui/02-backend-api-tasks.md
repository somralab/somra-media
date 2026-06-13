# Sprint 05 — Backend API Görevleri (Gezinme & Keşif)

> **Sprint hedefi:** Kütüphane gezinme, keşif rafları, arama ve filtreleme için verimli API'ler.
>
> **İlgili:** [`01-frontend-tasks.md`](./01-frontend-tasks.md) · Sprint 02 şema/FTS

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- Sprint 02 (medya/FTS), Sprint 03 (kullanıcı/izleme durumu/ebeveyn kontrolü).

## Epikler ve Görevler

### Epik A: Gezinme API'leri
- [ ] A1 — Sayfalı/filtreli kütüphane listesi endpoint'i | Kabul: hızlı, indeks kullanır.
- [ ] A2 — Öğe/sezon/bölüm detay endpoint'leri | Kabul: tek çağrıda gerekli veri.

### Epik B: Keşif rafları
- [ ] B1 — "Devam et", "yeni eklenenler", öneri rafı endpoint'leri (kullanıcı bazlı) | Kabul: izleme durumuna göre.
- [ ] B2 — Ebeveyn kontrolü filtre uygulaması (sunucu tarafı) | Kabul: kısıtlı içerik dönmez.

### Epik C: Arama
- [ ] C1 — FTS tabanlı arama endpoint'i (debounce dostu) | Kabul: düşük gecikme.

## Kabul Kriterleri (Sprint Çıktısı)
- Frontend gezinme/keşif/arama ihtiyaçlarını karşılayan performanslı API'ler hazır.

## Riskler
- Sunucu tarafı ebeveyn filtresi tutarlılığı → her endpoint'te uygulanmalı.

## Kapsam Dışı
- Akıllı öneri/ML — bu plan kapsamı dışında (basit kural tabanlı raflar).
