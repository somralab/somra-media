# Sprint 02 — Backend Görevleri (Kütüphane & Tarama)

> **Sprint hedefi:** Medya kütüphanesi tanımı, dosya tarama, dosya izleme (watch) ve
> ffprobe ile teknik metadata çıkarımı. M2'nin temeli.
>
> **İlgili:** [`../architecture.md`](../architecture.md) §3 · [`../project-brief.md`](../project-brief.md) · [`../definition-of-done.md`](../definition-of-done.md) · Sprint 01

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (gözetim)

## Bağımlılıklar
- Sprint 01: job scheduler, veri katmanı, API gateway.

## Epikler ve Görevler

### Epik A: Kütüphane tanımı
- [x] A1 — Kütüphane (library) kavramı: tür (film/dizi/müzik), kaynak klasör(ler), tarama ayarları | Kabul: CRUD API + test.
- [x] A2 — Çoklu klasör/volume desteği | Kabul: birden fazla yol taranabilir.

### Epik B: Dosya tarama motoru
- [x] B1 — Tam tarama (full scan) işi: dosya keşfi, desteklenen format filtresi | Kabul: büyük klasörde stabil çalışır, ilerleme raporlar.
- [x] B2 — Artımlı tarama (yalnızca değişenler) | Kabul: değişiklik tespiti çalışır.
- [x] B3 — ffprobe ile teknik metadata (kodek, çözünürlük, süre, ses/altyazı kanalları) | Kabul: doğru parse, test verisiyle doğrulanır.
- [x] B4 — Dosya adı/klasör yapısından ön ayrıştırma (başlık, yıl, sezon/bölüm) | Kabul: yaygın adlandırma kalıpları çözülür.

### Epik C: Dosya izleme (watch)
- [x] C1 — Dosya sistemi izleyici (ekleme/silme/taşıma) | Kabul: değişiklikte artımlı tarama tetiklenir.
- [x] C2 — Debounce/toplu işleme | Kabul: kütle değişimde sistem boğulmaz.

## Kabul Kriterleri (Sprint Çıktısı)
- Bir kütüphane tanımlanıp taranır; teknik metadata DB'ye yazılır; izleme aktif.
- Tüm işler scheduler üzerinden, ilerleme raporlu.

## Riskler
- Çeşitli adlandırma/format kombinasyonları → kapsamlı test verisi gerekir.

## Kapsam Dışı
- Zenginleştirilmiş (harici sağlayıcı) metadata — bkz. [`03-metadata-providers-tasks.md`](./03-metadata-providers-tasks.md).
