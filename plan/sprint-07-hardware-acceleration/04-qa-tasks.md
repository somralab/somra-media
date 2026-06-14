# Sprint 07 — QA Görevleri

> **Sprint hedefi:** Donanım hızlandırmanın doğruluğunu, performans kazancını ve fallback
> dayanıklılığını doğrulamak. M4 beta adayı kontrolü.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md) · [`../roadmap.md`](../roadmap.md) (M4)

## Sorumlu Rol(ler)
- QA (birincil), Medya Uzmanı (donanım ortamı)

## Bağımlılıklar
- Bu sprint medya/devops/backend çıktıları.

## Epikler ve Görevler

### Epik A: İşlevsel
- [x] A1 — HW transcode doğruluğu (görüntü/ses kalitesi) | Kabul: çıktı kabul edilebilir.
- [x] A2 — Otomatik seçim ve HW→SW fallback testi | Kabul: kesintisiz geçiş.

### Epik B: Performans
- [x] B1 — CPU vs HW karşılaştırma (kaynak/eşzamanlı oturum) | Kabul: belirgin kazanç raporlanır.
- [x] B2 — Donanım oturum limiti stres testi | Kabul: limit korunur.

### Epik C: Beta kabul
- [x] C1 — M4 beta adayı kontrol listesi | Kabul: beta kriterleri karşılanır.

## Kabul Kriterleri (Sprint Çıktısı)
- HW yolu test kapsamında; performans kazancı ölçülü; fallback güvenli; M4 kriterleri sağlanır.

## Riskler
- Donanım çeşitliliği test ortamını sınırlar → öncelikli donanımda derin test.

## Kapsam Dışı
- Tüm GPU modellerinde sertifikasyon — best-effort.
