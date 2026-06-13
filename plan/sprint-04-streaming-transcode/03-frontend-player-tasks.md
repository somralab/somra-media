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
- [ ] A1 — hls.js entegrasyonu + manifest yükleme | Kabul: video oynar.
- [ ] A2 — Oynat/duraklat/seek/ses kontrolü + klavye kısayolları | Kabul: temel kontroller çalışır.
- [ ] A3 — Tam ekran + responsive davranış | Kabul: masaüstü/mobil tarayıcıda düzgün.

### Epik B: Akış seçenekleri
- [ ] B1 — Kalite (ABR) seçimi (otomatik + manuel) | Kabul: kademe değişimi sorunsuz.
- [ ] B2 — Ses dili ve altyazı seçimi UI'si | Kabul: backend akışlarına bağlı.

### Epik C: Devam etme
- [ ] C1 — Resume (kaldığı yerden) + periyodik ilerleme bildirimi | Kabul: doğru pozisyon.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı tarayıcıda video izler; kalite/ses/altyazı seçer; kaldığı yerden devam eder.

## Riskler
- Tarayıcı/kodek uyumluluğu → test matrisi (Chrome/Firefox/Safari).

## Kapsam Dışı
- Zengin kütüphane gezinme ekranları — Sprint 05.
