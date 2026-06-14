# Sprint 02 — Frontend Görevleri (Kütüphane Yönetimi)

> **Sprint hedefi:** Kütüphane tanımlama, tarama tetikleme ve tarama/metadata durumunu
> gösteren yönetim arayüzü (henüz oynatma yok).
>
> **İlgili:** Sprint 01 [`../sprint-01-foundation/04-frontend-tasks.md`](../sprint-01-foundation/04-frontend-tasks.md) · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Sorumlu Rol(ler)
- Frontend (birincil)

## Bağımlılıklar
- Sprint 01 SPA iskeleti + API istemci; bu sprint backend kütüphane/tarama API'leri.

## Epikler ve Görevler

### Epik A: Kütüphane yönetimi
- [x] A1 — Kütüphane oluştur/düzenle/sil ekranları | Kabul: CRUD API'ye bağlı, doğrulamalı.
- [x] A2 — Klasör/yol seçimi UI'si | Kabul: çoklu yol eklenebilir.

### Epik B: Tarama izleme
- [x] B1 — Tarama tetikleme + gerçek zamanlı ilerleme (WS/SSE) | Kabul: ilerleme canlı güncellenir.
- [x] B2 — Tarama geçmişi / iş durumu görünümü | Kabul: başarılı/hatalı işler listelenir.

### Epik C: Metadata önizleme
- [x] C1 — Eşleşen öğelerin temel listesi + metadata önizleme | Kabul: poster + başlık + yıl gösterilir.
- [x] C2 — Manuel yeniden eşleştirme arayüzü | Kabul: backend düzeltme API'sini kullanır.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı arayüzden kütüphane tanımlar, tarar ve sonuçları görür.

## Riskler
- Gerçek zamanlı ilerleme UX'i → WS/SSE entegrasyonu sağlam olmalı.

## Kapsam Dışı
- Zengin gezinme/oynatıcı ekranları — Sprint 05.
