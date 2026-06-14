# Sprint 08 — Backend Görevleri (İstek Yönetimi — Overseerr İşlevi)

> **Sprint hedefi:** Kullanıcıların olmayan içerik için istek oluşturması, onay akışı ve
> istek durumu takibi (Overseerr/Jellyseerr işlevi).
>
> **İlgili:** [`../architecture.md`](../architecture.md) (İstek Yönetimi) · Sprint 02 (metadata arama) · Sprint 03 (kullanıcı/RBAC)

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- Sprint 02 (metadata sağlayıcı arama), Sprint 03 (kullanıcı/yetki).

## Epikler ve Görevler

### Epik A: İstek modeli
- [x] A1 — `request` şeması (film/dizi, istek sahibi, durum, çözünürlük/kalite tercihi) | Kabul: migrasyon + CRUD.
- [x] A2 — Mevcut kütüphaneyle çakışma kontrolü (zaten var mı) | Kabul: tekrar istek engellenir/işaretlenir.

### Epik B: Onay akışı
- [x] B1 — İstek durum makinesi (beklemede → onaylandı/reddedildi → tamamlandı) | Kabul: durum geçişleri kontrollü.
- [x] B2 — Rol bazlı otomatik onay/kota (admin politikası) | Kabul: kullanıcı kotası uygulanır.

### Epik C: Keşif & arama
- [x] C1 — Sağlayıcı üzerinden "eklenebilir içerik" arama (kütüphanede olmayan) | Kabul: arama sonuç döner.
- [x] C2 — Otomasyon köprüsü iskeleti (Sprint 09 ile bağlanacak handoff noktası) | Kabul: onaylı istek otomasyona devredilebilir arayüz.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı içerik isteyebilir; admin onaylar; istek durumu izlenir; bildirim tetiklenir.

## Riskler
- Sprint 09 otomasyonuyla entegrasyon → handoff arayüzü net tanımlanmalı.

## Kapsam Dışı
- Gerçek indirme/kapma — Sprint 09.
