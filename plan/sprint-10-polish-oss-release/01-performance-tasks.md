# Sprint 10 — Performans & Sertleştirme Görevleri

> **Sprint hedefi:** Tüm sistemi performans, kararlılık ve kaynak verimliliği açısından
> sertleştirmek. M6 (1.0) için kalite bariyeri.
>
> **İlgili:** [`../project-brief.md`](../project-brief.md) (başarı kriterleri: ev donanımında verim) · [`../roadmap.md`](../roadmap.md) (M6)

## Sorumlu Rol(ler)
- Tech Lead (birincil), Backend, Medya Uzmanı, Frontend

## Bağımlılıklar
- Tüm önceki sprintlerin çıktıları.

## Epikler ve Görevler

### Epik A: Backend performans
- [ ] A1 — Profilleme (CPU/bellek) + sıcak yol optimizasyonu | Kabul: ölçülen iyileşme.
- [ ] A2 — DB sorgu/indeks optimizasyonu (büyük kütüphane) | Kabul: yavaş sorgular giderilir.
- [ ] A3 — Tarama/metadata/otomasyon işlerinin kaynak verimi | Kabul: ev donanımında stabil.

### Epik B: Streaming verimi
- [ ] B1 — Transcode/HW oturum verimliliği ve gecikme | Kabul: hedef oturum sayısı karşılanır.

### Epik C: Frontend performans
- [ ] C1 — Bundle boyutu, lazy load, render optimizasyonu | Kabul: yükleme/etkileşim metrikleri iyi.

### Epik D: Kararlılık
- [ ] D1 — Uzun süre çalışma (soak) testi + bellek sızıntısı kontrolü | Kabul: sızıntı yok.

## Kabul Kriterleri (Sprint Çıktısı)
- Sistem ev sunucusu donanımında (ör. 4 çekirdek/8GB) hedeflenen yükte stabil ve verimli çalışır.

## Riskler
- Geç bulunan darboğazlar → erken profilleme alışkanlığı.

## Kapsam Dışı
- Yeni özellik geliştirme — bu sprint cila odaklıdır.
