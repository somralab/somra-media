# Sprint 05 — QA Görevleri

> **Sprint hedefi:** Uçtan uca kullanıcı akışını (giriş → gezinme → arama → oynatma) ve M3
> alfa kalitesini doğrulamak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) · [`../roadmap.md`](../roadmap.md) (M3)

## Sorumlu Rol(ler)
- QA (birincil)

## Bağımlılıklar
- Bu sprint frontend/backend + Sprint 04 oynatıcı.

## Epikler ve Görevler

### Epik A: E2E akışlar
- [x] A1 — Giriş → kütüphane → detay → oynatma e2e | Kabul: kritik yol geçer.
- [x] A2 — Arama/filtre/raf doğruluğu | Kabul: sonuçlar tutarlı.

### Epik B: Uyumluluk & erişilebilirlik
- [x] B1 — Responsive/tarayıcı testi | Kabul: masaüstü/mobil tarayıcı.
- [x] B2 — Erişilebilirlik kontrolü (klavye, kontrast) | Kabul: temel WCAG (her dört temada).
- [x] B3 — Tema testi: dört temada tutarlılık + tema kalıcılığı | Kabul: tema değişir, hatırlanır, kontrast korunur.

### Epik C: Alfa kabul
- [x] C1 — M3 alfa kabul kontrol listesi | Kabul: alfa demoya hazır.

## Kabul Kriterleri (Sprint Çıktısı)
- Uçtan uca akış test kapsamında; M3 alfa kriterleri karşılanır.

## Riskler
- Performans regresyonu büyük kütüphanede → ölçümlü test.

## Kapsam Dışı
- Onboarding sihirbazı testi — Sprint 06.
