# Sprint 03 — QA Görevleri

> **Sprint hedefi:** Kimlik, RBAC, ebeveyn kontrolü ve izleme durumu akışlarını doğrulamak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`04-security-tasks.md`](./04-security-tasks.md)

## Sorumlu Rol(ler)
- QA (birincil)

## Bağımlılıklar
- Bu sprint backend + frontend çıktıları.

## Epikler ve Görevler

### Epik A: İşlevsel testler
- [x] A1 — Giriş/çıkış/yenileme e2e akışı | Kabul: kritik yollar geçer.
- [x] A2 — RBAC matris testi (her rol için erişim doğrulama) | Kabul: yetki ihlali yok.
- [x] A3 — Ebeveyn kontrolü testi (çocuk profili kısıtları) | Kabul: kısıtlı içerik görünmez.

### Epik B: İzleme durumu
- [x] B1 — Devam etme (resume) ve izlendi durumu testi | Kabul: doğru pozisyon korunur.

### Epik C: Güvenlik kabul testleri
- [x] C1 — Brute-force/rate-limit ve yetkisiz erişim testleri | Kabul: korumalar tetiklenir.

## Kabul Kriterleri (Sprint Çıktısı)
- Kimlik/RBAC/ebeveyn akışları test kapsamında; kritik/güvenlik hatası yok.

## Riskler
- Yetki edge-case'leri → matris testi kapsamlı olmalı.

## Kapsam Dışı
- Oynatma performansı — Sprint 04.
