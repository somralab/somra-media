# Sprint 01 — Backend Görevleri

> **Sprint hedefi:** Çekirdek servis iskeletinin backend bileşenleri: uygulama yaşam döngüsü,
> job scheduler iskeleti, ortak yardımcılar.
>
> **İlgili:** [`../architecture.md`](../architecture.md) · [`../tech-stack.md`](../tech-stack.md) · [`../definition-of-done.md`](../definition-of-done.md) · [`01-architecture-tasks.md`](./01-architecture-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (gözetim)

## Bağımlılıklar
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) Epik A/B (modül sınırları, bootstrap).

## Epikler ve Görevler

### Epik A: Uygulama çekirdeği
- [x] A1 — Uygulama başlatma/kapatma yaşam döngüsü (graceful shutdown, signal handling) | Kabul: SIGTERM'de temiz kapanır, test edilir.
- [x] A2 — Konfigürasyon okuma + doğrulama (env + varsayılan) | Kabul: hatalı konfigde anlamlı hata.
- [x] A3 — Ortak hata tipleri ve sarmalama yardımcıları | Kabul: standart hata yanıtı formatı.

### Epik B: Job Scheduler iskeleti
- [x] B1 — Periyodik + tek-seferlik iş çalıştırma altyapısı | Kabul: örnek job çalışır, loglanır.
- [x] B2 — İş durumu izleme (çalışıyor/başarılı/hata) ve eşzamanlılık koruması | Kabul: aynı işin çakışması engellenir.
- [x] B3 — İş kuyruğu API iskeleti (sonraki sprintlerde tarama/yenileme bağlanacak) | Kabul: arayüz sözleşmesi tanımlı.

### Epik C: Ortak altyapı
- [x] C1 — Yapılandırılmış logger entegrasyonu | Kabul: tüm modüllerde tutarlı.
- [x] C2 — Sağlık/teşhis (diagnostics) bilgisi toplama iskeleti | Kabul: `/api/v1/health` zenginleştirilir.
- [x] C3 — Backend i18n iskeleti: locale-aware mesaj kataloğu + dil pazarlığı (kullanıcı/sistem/`Accept-Language`/en-US) + API hata yanıtında anahtar+yerelleştirilmiş mesaj | Kabul: örnek hata mesajı en-US/tr-TR döner. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §4.3.

## Kabul Kriterleri (Sprint Çıktısı)
- Scheduler örnek bir periyodik işi güvenle çalıştırır.
- Tüm backend kodu DoD §3 standartlarına uyar; testler yeşil.

## Riskler
- Scheduler tasarımı sonraki sprintlerin tüm asenkron işlerini taşıyacak → arayüz erken sağlam tanımlanmalı.

## Kapsam Dışı
- Gerçek iş mantığı (tarama, metadata) — Sprint 02.
