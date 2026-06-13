# Sprint 04 — Backend Görevleri (Oynatma Entegrasyonu)

> **Sprint hedefi:** Streaming pipeline'ını kullanıcı/izleme durumu ve API ile bütünleştirmek.
>
> **İlgili:** [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md) · Sprint 03 izleme durumu

## Sorumlu Rol(ler)
- Backend (birincil), Medya Uzmanı (pipeline arayüzü)

## Bağımlılıklar
- [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md), Sprint 03 (`watch_state`, yetki).

## Epikler ve Görevler

### Epik A: Oynatma API'si
- [ ] A1 — "Oynat" endpoint'i: yetki + oynatma kararı + oturum başlatma | Kabul: uçtan uca akış başlar.
- [ ] A2 — İzleme ilerleme güncelleme (periyodik ping) | Kabul: resume verisi güncellenir.

### Epik B: Oturum & kaynak yönetimi
- [ ] B1 — Eşzamanlı transcode oturum limiti + kuyruğa alma | Kabul: aşırı yük engellenir.
- [ ] B2 — Boşta kalan oturum sonlandırma | Kabul: kaynak sızıntısı yok.

### Epik C: Telemetri
- [ ] C1 — Oynatma/transcode metrikleri (oturum sayısı, hata oranı) | Kabul: temel metrik toplanır.

## Kabul Kriterleri (Sprint Çıktısı)
- Oynatma API'si yetki + izleme durumu + oturum yönetimiyle çalışır.

## Riskler
- Eşzamanlı oturum yönetimi ev donanımında kritik → limit/kuyruk şart.

## Kapsam Dışı
- Donanım hızlandırma seçimi — Sprint 07.
