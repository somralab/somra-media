# Sprint 08 — Frontend Görevleri (İstek Arayüzü)

> **Sprint hedefi:** İçerik keşfi/istek oluşturma, istek takibi ve admin onay arayüzleri.
>
> **İlgili:** [`01-backend-tasks.md`](./01-backend-tasks.md) · Sprint 05 (gezinme UI deseni)

## Sorumlu Rol(ler)
- Frontend (birincil)

## Bağımlılıklar
- [`01-backend-tasks.md`](./01-backend-tasks.md) istek API'leri.

## Epikler ve Görevler

### Epik A: İstek oluşturma
- [ ] A1 — Keşif/arama (kütüphanede olmayan içerik) + "iste" akışı | Kabul: kullanıcı istek oluşturur.
- [ ] A2 — Kalite/çözünürlük tercihi seçimi | Kabul: backend'e iletilir.

### Epik B: İstek takibi
- [ ] B1 — "İsteklerim" ekranı + durum gösterimi | Kabul: gerçek zamanlı durum.

### Epik C: Admin onayı
- [ ] C1 — Bekleyen istekler + onay/ret arayüzü | Kabul: durum güncellenir.
- [ ] C2 — Kota/politika ayar arayüzü | Kabul: backend politikasına bağlı.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı istek oluşturup takip eder; admin onaylar/reddeder.

## Riskler
- Durum senkronizasyonu → WS/SSE ile canlı güncelleme.

## Kapsam Dışı
- Otomasyon ayar ekranları — Sprint 09.
