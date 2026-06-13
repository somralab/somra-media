# Sprint 01 — QA Görevleri

> **Sprint hedefi:** Test stratejisini kurmak, otomasyon iskeletini hazırlamak ve DoD
> doğrulama sürecini başlatmak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) §4 · [`../project-brief.md`](../project-brief.md)

## Sorumlu Rol(ler)
- QA (birincil), tüm geliştiriciler (test yazımı)

## Bağımlılıklar
- CI hattı ([`05-devops-tasks.md`](./05-devops-tasks.md) Epik B).

## Epikler ve Görevler

### Epik A: Test stratejisi
- [ ] A1 — Test piramidi ve kapsam politikası dokümante | Kabul: birim/entegrasyon/e2e sınırları net.
- [ ] A2 — Hata/issue yönetim süreci ve önem seviyeleri | Kabul: kritik/yüksek/orta/düşük tanımlı.
- [ ] A3 — Coverage standardının ([`../definition-of-done.md`](../definition-of-done.md) §4.1) operasyonel tanımı: ölçüm aracı, kritik modül listesi, rapor formatı | Kabul: çekirdek ≥%80, kritik ≥%90, frontend bileşen ≥%70 eşikleri uygulanabilir.

### Epik B: Otomasyon iskeleti
- [ ] B1 — Backend entegrasyon test çatısı (izole DB ile) | Kabul: örnek test CI'da koşar.
- [ ] B2 — E2E test çatısı kurulumu (web akışları için) | Kabul: `health` sayfası smoke testi geçer.

### Epik C: DoD doğrulama
- [ ] C1 — Sprint kapanış kontrol listesi (DoD §1–§2) | Kabul: her sprintte uygulanır.
- [ ] C2 — i18n kabul ölçütü kontrol listesi (hardcoded metin yok, en-US+tr-TR tam) | Kabul: her sprintte i18n doğrulanır. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §6.

## Kabul Kriterleri (Sprint Çıktısı)
- Test çatıları CI'da çalışır; smoke testleri yeşil.
- Test stratejisi ve hata süreci dokümante.

## Riskler
- Erken test altyapısı eksikliği teknik borç biriktirir → bu sprintte temel atılır.

## Kapsam Dışı
- Özellik bazlı kapsamlı test senaryoları — ilgili sprintlerde.
