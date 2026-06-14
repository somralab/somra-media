# Sprint 04 — Frontend Görevleri (Web Oynatıcı)

> **Sprint hedefi:** hls.js tabanlı web video oynatıcı: oynatma, seek, kalite/ses/altyazı seçimi,
> devam etme.
>
> **İlgili:** [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md) · [`02-backend-tasks.md`](./02-backend-tasks.md) · [`../tech-stack.md`](../tech-stack.md)

## Sorumlu Rol(ler)
- Frontend (birincil), Medya Uzmanı (uyumluluk desteği)

## Bağımlılıklar
- Backend streaming/oynatma API'leri.

## Epikler ve Görevler

### Epik A: Oynatıcı çekirdeği
- [x] A1 — hls.js entegrasyonu + manifest yükleme | Kabul: video oynar.
- [x] A2 — Oynat/duraklat/seek/ses kontrolü + klavye kısayolları | Kabul: temel kontroller çalışır.
- [x] A3 — Tam ekran + responsive davranış | Kabul: masaüstü/mobil tarayıcıda düzgün.

### Epik B: Akış seçenekleri
- [x] B1 — Kalite (ABR) seçimi (otomatik + manuel) | Kabul: kademe değişimi sorunsuz.
- [x] B2 — Ses dili ve altyazı seçimi UI'si | Kabul: backend akışlarına bağlı.

### Epik C: Devam etme
- [x] C1 — Resume (kaldığı yerden) + periyodik ilerleme bildirimi | Kabul: doğru pozisyon.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı tarayıcıda video izler; kalite/ses/altyazı seçer; kaldığı yerden devam eder.

## Riskler
- Tarayıcı/kodek uyumluluğu → test matrisi (Chrome/Firefox/Safari).

## Kapsam Dışı
- Zengin kütüphane gezinme ekranları — Sprint 05.
