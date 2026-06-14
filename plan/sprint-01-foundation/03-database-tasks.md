# Sprint 01 — Veritabanı Görevleri

> **Sprint hedefi:** SQLite veri katmanı temeli, migrasyon altyapısı ve şema sürümleme.
>
> **İlgili:** [`../tech-stack.md`](../tech-stack.md) · [`../architecture.md`](../architecture.md) §4 · [`../definition-of-done.md`](../definition-of-done.md)

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (şema gözden geçirme)

## Bağımlılıklar
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) Epik A3 (migrasyon aracı kararı).

## Epikler ve Görevler

### Epik A: Veri erişim katmanı
- [x] A1 — SQLite bağlantı yönetimi (WAL modu, bağlantı havuzu, pragma ayarları) | Kabul: WAL aktif, eşzamanlı okuma testi geçer.
- [x] A2 — Repository/erişim deseni iskeleti | Kabul: örnek tablo için CRUD testli.
- [x] A3 — İşlem (transaction) yardımcıları | Kabul: rollback/commit testleri.

### Epik B: Migrasyon ve şema sürümleme
- [x] B1 — Migrasyon altyapısı (ileri/geri) kurulumu | Kabul: `up`/`down` çalışır, sürüm tablosu tutulur.
- [x] B2 — Şema sürümleme ve uygulama açılışında otomatik migrasyon | Kabul: yükseltmede şema otomatik güncellenir.
- [x] B3 — Tohum (seed) ve test veri altyapısı | Kabul: test DB izole kurulur.

### Epik C: Yedekleme/dayanıklılık temeli
- [x] C1 — DB dosyası konumu ve volume stratejisi (kalıcılık) | Kabul: [`05-devops-tasks.md`](./05-devops-tasks.md) ile uyumlu volume.
- [x] C2 — Bütünlük kontrolü ve bozulma kurtarma notları | Kabul: temel `PRAGMA integrity_check` akışı.

## Kabul Kriterleri (Sprint Çıktısı)
- Uygulama açılışta migrasyonları uygular; örnek repository testleri geçer.
- Veri katmanı [`../architecture.md`](../architecture.md) §4 ile uyumludur.

## Riskler
- Şema kararları erken; sonraki sprintlerde sık migrasyon olabilir → migrasyon disiplini şart.

## Kapsam Dışı
- Domain tabloları (medya, kullanıcı vb.) — ilgili sprintlerde eklenir.
