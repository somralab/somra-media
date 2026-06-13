# Sprint 10 — QA Görevleri (Yayın Kabulü)

> **Sprint hedefi:** 1.0 yayını için kapsamlı regresyon, kabul ve yayın kalite kapısı.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) · [`../roadmap.md`](../roadmap.md) (M6) · [`../project-brief.md`](../project-brief.md) (başarı kriterleri)

## Sorumlu Rol(ler)
- QA (birincil), tüm ekip

## Bağımlılıklar
- Tüm önceki sprintler.

## Epikler ve Görevler

### Epik A: Tam regresyon
- [ ] A1 — Tüm sprint regresyon paketlerinin birleşik koşumu | Kabul: yeşil.
- [ ] A2 — Uçtan uca senaryo matrisi (kurulum → kullanım → otomasyon) | Kabul: kritik akışlar geçer.

### Epik B: Yayın kabul kriterleri
- [ ] B1 — Brief başarı kriterleri doğrulaması (kurulum süresi, optimizasyon, bütünlük, performans) | Kabul: tüm kriterler karşılanır.
- [ ] B2 — Çok platform/tarayıcı son kontrol | Kabul: hedef ortamlar çalışır.
- [ ] B3 — Yükseltme/geri yükleme kabul testi | Kabul: veri korunur.
- [ ] B4 — i18n yayın kapısı: en-US + tr-TR %100 anahtar tamlığı, hardcoded metin taraması, pseudo-locale taşma/uzunluk testi | Kabul: iki dil eksiksiz, taşma yok. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §6.

### Epik C: Yayın kapısı
- [ ] C1 — 1.0 yayın kontrol listesi (kod, doküman, güvenlik, imaj) | Kabul: tüm maddeler tamam.

## Kabul Kriterleri (Sprint Çıktısı)
- M6 (1.0) yayın kalite kapısı geçilir; bilinen kritik/yüksek hata yok; ürün yayına hazır.

## Riskler
- Son aşama hata yığılması → sprint boyunca sürekli regresyon.

## Kapsam Dışı
- Yayın sonrası bakım/2.0 planlaması — ayrı planlanır.
