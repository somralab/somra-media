# Sprint 01 — Mimari Görevleri

> **Sprint hedefi:** Çalışan bir iskelet servis, net modül sınırları, API sözleşmesi ve
> CI/CD temeli. Bu sprint sonunda `docker run` ile boş ama ayağa kalkan bir Somra çekirdeği olur (M1).
>
> **İlgili:** [`../project-brief.md`](../project-brief.md) · [`../architecture.md`](../architecture.md) · [`../tech-stack.md`](../tech-stack.md) · [`../definition-of-done.md`](../definition-of-done.md) · [`../i18n-localization.md`](../i18n-localization.md)

## Sorumlu Rol(ler)
- Tech Lead / Mimar (birincil), Backend (destek)

## Bağımlılıklar
- Yok (başlangıç sprinti). Çıktıları **tüm** sonraki sprintlerin temelidir.

## Epikler ve Görevler

### Epik A: Kapatılan kararların uygulanması ve doğrulanması
> Tüm teknoloji/mimari kararlar verildi (bkz. [`../tech-stack.md`](../tech-stack.md) §7, [`../architecture.md`](../architecture.md) §8). Bu epik kararları **uygular/doğrular**, yeniden tartışmaz.
- [ ] A1 — HTTP router (`go-chi/chi`) entegrasyonu + iskelet | Kabul: yönlendirme + middleware çalışır.
- [ ] A2 — Oturum/kimlik temeli (JWT erişim + iptal edilebilir refresh token) iskeleti | Kabul: Sprint 03 için sözleşme hazır.
- [ ] A3 — Migrasyon (`pressly/goose`) + scheduler (`robfig/cron/v3`) entegrasyonu | Kabul: örnek migrasyon + cron işi çalışır.
- [ ] A4 — OpenAPI 3.1 design-first sözleşme iskeleti + FE tip üretim hattı | Kabul: `/api/v1/health` spec'ten tip üretilir.
- [ ] A5 — i18n mimarisi uygulaması: kütüphaneler (`i18next`/`react-i18next`, `go-i18n/v2`), anahtar standardı (`domain.context.key`), dil pazarlığı | Kabul: çalışan iskelet. Bkz. [`../i18n-localization.md`](../i18n-localization.md).

### Epik B: Modül iskeleti
- [ ] B1 — Monorepo dizin yapısı ve modül sınırları (API, kimlik, kütüphane, metadata, streaming, ayarlar, jobs) | Kabul: [`../architecture.md`](../architecture.md) §3 ile birebir örtüşür.
- [ ] B2 — Bağımlılık enjeksiyonu / uygulama başlatma (bootstrap) iskeleti | Kabul: servis temiz başlar/kapanır (graceful shutdown).
- [ ] B3 — Konfigürasyon katmanı (ortam değişkenleri + varsayılanlar) | Kabul: "convention over configuration" ilkesi uygulanır.
- [ ] B4 — Yapılandırılmış loglama + hata yönetimi standardı | Kabul: tüm modüller ortak logger kullanır.

### Epik C: API Gateway temeli
- [ ] C1 — HTTP sunucu + `/api/v1/health` ve `/api/v1/version` uçları | Kabul: 200 döner, testi var.
- [ ] C2 — Middleware zinciri (request log, recover, CORS, rate-limit iskeleti) | Kabul: birim testleriyle doğrulanır.
- [ ] C3 — WebSocket/SSE altyapı iskeleti (gerçek zamanlı olaylar için) | Kabul: örnek olay yayını çalışır.

## Kabul Kriterleri (Sprint Çıktısı)
- Servis tek binary olarak derlenir ve ayağa kalkar; health/version uçları yanıt verir.
- Mimari kararlar dokümanlara işlenmiştir; açık karar listesi kapanmıştır.
- [`../definition-of-done.md`](../definition-of-done.md) §1–§2 karşılanır.

## Riskler
- Erken yanlış mimari karar maliyeti yüksek → kararlar dokümante edilip gözden geçirilir.

## Kapsam Dışı
- İş mantığı (tarama, oynatma vb.) — sonraki sprintler. Sadece iskelet.
