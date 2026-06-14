# Sprint 08 — QA Görevleri

> **Sprint hedefi:** İstek akışı, onay ve bildirimleri doğrulamak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Sorumlu Rol(ler)
- QA (birincil)

## Bağımlılıklar
- Bu sprint backend/frontend/bildirim çıktıları.

## Epikler ve Görevler

### Epik A: İstek akışı
- [x] A1 — İstek oluştur → onay/ret → durum e2e | Kabul: tüm geçişler doğru.
- [x] A2 — Çakışma/kota/yetki testleri | Kabul: politikalar uygulanır.

### Epik B: Bildirim
- [x] B1 — Her kanal için tetikleme testi | Kabul: doğru olayda iletilir.
- [x] B2 — Tercih/abonelik testi | Kabul: istenmeyen bildirim gitmez.

### Epik C: Regresyon
- [x] C1 — Sprint 08 regresyon paketi | Kabul: CI'da koşar.

## Kabul Kriterleri (Sprint Çıktısı)
- İstek ve bildirim akışları test kapsamında; kritik hata yok.

## Riskler
- Harici kanal bağımlılığı → mock + sınırlı gerçek test.

## Kapsam Dışı
- Otomasyon/indexer testleri — Sprint 09.
