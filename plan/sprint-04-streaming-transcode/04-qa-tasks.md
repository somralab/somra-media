# Sprint 04 — QA Görevleri

> **Sprint hedefi:** Oynatma ve transcode doğruluğunu, uyumluluğunu ve dayanıklılığını doğrulamak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md)

## Sorumlu Rol(ler)
- QA (birincil), Medya Uzmanı (test medyası)

## Bağımlılıklar
- Bu sprint streaming + oynatıcı çıktıları.

## Epikler ve Görevler

### Epik A: Uyumluluk matrisi
- [x] A1 — Kodek/konteyner test seti (H.264/H.265/AAC/AC3/MKV/MP4 vb.) | Kabul: matris sonucu raporlanır.
- [x] A2 — Tarayıcı uyumluluk testi (Chrome/Firefox/Safari) | Kabul: kritik formatlar oynar.

### Epik B: İşlevsel testler
- [x] B1 — Direct play vs transcode karar doğruluğu | Kabul: gereksiz transcode olmaz.
- [x] B2 — Seek/altyazı/ses seçimi/resume e2e | Kabul: akışlar geçer.

### Epik C: Dayanıklılık & performans
- [x] C1 — Eşzamanlı oturum/limit stres testi | Kabul: limit korunur, çökmez.
- [x] C2 — Transcode kaynak/temizlik testi | Kabul: süreç/disk sızıntısı yok.

## Kabul Kriterleri (Sprint Çıktısı)
- Oynatma akışları test kapsamında; uyumluluk matrisi belgeli; kritik hata yok.

## Riskler
- Geniş format çeşitliliği → otomasyon + örnek medya havuzu gerekir.

## Kapsam Dışı
- Donanım hızlandırma testleri — Sprint 07.
