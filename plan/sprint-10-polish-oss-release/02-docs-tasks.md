# Sprint 10 — Dokümantasyon & Topluluk Görevleri

> **Sprint hedefi:** Açık kaynak yayını için eksiksiz kullanıcı/geliştirici dokümantasyonu ve
> topluluk altyapısı.
>
> **İlgili:** [`../project-brief.md`](../project-brief.md) §5 (lisans) · [`03-devops-release-tasks.md`](./03-devops-release-tasks.md)

## Sorumlu Rol(ler)
- PM (birincil), Tech Lead, tüm ekip (katkı)

## Bağımlılıklar
- Stabil özellik seti (Sprint 01–09).

## Epikler ve Görevler

### Epik A: Kullanıcı dokümantasyonu (TR + EN)
- [ ] A1 — Kurulum rehberi (Docker/compose, GPU passthrough) | Kabul: sıfır bilgiyle kurulabilir; **tr-TR + en-US**.
- [ ] A2 — Özellik/kullanım kılavuzları (kütüphane, oynatma, istek, otomasyon) | Kabul: ana akışlar belgeli; **tr-TR + en-US**.
- [ ] A3 — SSS + sorun giderme | Kabul: yaygın sorunlar kapsanır; **tr-TR + en-US**.
- [ ] A4 — Doküman çeviri tamlığı (TR↔EN paritesi) | Kabul: iki dil de eksiksiz. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §2.

### Epik B: Geliştirici dokümantasyonu
- [ ] B1 — Mimari/katkı rehberi (CONTRIBUTING) + kod standartları | Kabul: [`../definition-of-done.md`](../definition-of-done.md) ile hizalı.
- [ ] B2 — API dokümantasyonu (OpenAPI yayını) | Kabul: güncel sözleşme.
- [ ] B3 — Eklenti geliştirme rehberi | Kabul: 3. taraf eklenti yazılabilir.

### Epik C: Topluluk & lisans
- [ ] C1 — LICENSE (AGPL-3.0 onaylı), Davranış Kuralları, issue/PR şablonları | Kabul: repo OSS standartlarına uygun.
- [ ] C2 — README + proje tanıtımı (somra markası) | Kabul: net değer önerisi.
- [ ] C3 — Çeviri katkı rehberi + **Weblate** kurulumu (self-host, git entegre) | Kabul: topluluk Weblate üzerinden çeviri katkısı yapabilir; repo senkronu çalışır. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §8.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı ve geliştirici dokümantasyonu eksiksiz; lisans ve topluluk altyapısı hazır.

## Riskler
- Eksik doküman benimsemeyi düşürür → yayından önce tamamlanmalı.

## Kapsam Dışı
- TR/EN dışındaki dillere çeviri — gelecekte (altyapı yeni dil eklemeye hazır olacak; bkz. [`../i18n-localization.md`](../i18n-localization.md) §1).
