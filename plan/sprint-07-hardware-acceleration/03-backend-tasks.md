# Sprint 07 — Backend Görevleri (HW Ayar & Entegrasyon)

> **Sprint hedefi:** Donanım hızlandırma ayarlarını sistem ayarları ve onboarding ile
> bütünleştirmek; izleme/telemetri.
>
> **İlgili:** Sprint 06 (akıllı varsayılan/ayar) · [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil), Medya Uzmanı (parametreler)

## Bağımlılıklar
- [`01-media-streaming-tasks.md`](./01-media-streaming-tasks.md), Sprint 06 ayar katmanı.

## Epikler ve Görevler

### Epik A: Ayar entegrasyonu
- [ ] A1 — HW hızlandırma ayarı (otomatik/zorla aç/kapat, hızlandırıcı seçimi) | Kabul: ayar API'de.
- [ ] A2 — Onboarding'e GPU tespit/öneri adımının eklenmesi | Kabul: sihirbaz GPU önerir.

### Epik B: İzleme
- [ ] B1 — HW oturum metrikleri (kullanım, fallback oranı) | Kabul: telemetri toplanır.
- [ ] B2 — HW hata/fallback loglama ve teşhis | Kabul: sorun teşhisi kolaylaşır.

## Kabul Kriterleri (Sprint Çıktısı)
- HW hızlandırma arayüzden yönetilebilir; tespit/öneri onboarding'e entegre; telemetri var.

## Riskler
- Yanlış ayar oynatmayı bozabilir → güvenli varsayılan "otomatik + fallback".

## Kapsam Dışı
- Yeni kodek araştırması (AV1 vb.) — gelecekteki yol haritası.
