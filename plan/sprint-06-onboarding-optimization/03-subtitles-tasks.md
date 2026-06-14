# Sprint 06 — Altyazı Otomasyonu Görevleri (Bazarr İşlevi)

> **Sprint hedefi:** Otomatik altyazı arama/indirme (Bazarr benzeri): eksik altyazı tespiti,
> sağlayıcı entegrasyonu ve eşleştirme.
>
> **İlgili:** [`../architecture.md`](../architecture.md) (Metadata/eklenti) · Sprint 04 (altyazı oynatma) · [`../project-brief.md`](../project-brief.md) (kapsam)

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- Sprint 02 (kütüphane/öğe), Sprint 04 (altyazı işleme).

## Epikler ve Görevler

### Epik A: Altyazı sağlayıcı entegrasyonu
- [x] A1 — Açık altyazı sağlayıcı(ları) için ortak arayüz + entegrasyon | Kabul: arama/indirme çalışır.
- [x] A2 — Dil tercihi ve kalite/eşleşme skorlama | Kabul: doğru altyazı seçilir.

### Epik B: Otomasyon
- [x] B1 — Eksik altyazı tespiti (kullanıcı dil tercihine göre) | Kabul: eksikler raporlanır.
- [x] B2 — Periyodik otomatik indirme işi (scheduler) | Kabul: yeni içerikte altyazı otomatik gelir.
- [x] B3 — Manuel altyazı arama/yükleme | Kabul: kullanıcı override edebilir.

### Epik C: UI bağlantısı
- [x] C1 — Detay sayfasında altyazı yönetimi (frontend ile koordine) | Kabul: altyazı durumu görünür.

## Kabul Kriterleri (Sprint Çıktısı)
- Sistem eksik altyazıları tespit edip otomatik indirir; manuel yönetim mümkün.

## Riskler
- Sağlayıcı oran sınırı/lisans → cache + uyumlu sağlayıcı seçimi.

## Kapsam Dışı
- Altyazı senkron/AI çeviri — bu plan kapsamı dışında.
