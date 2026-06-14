# Sprint 01 — Frontend Görevleri

> **Sprint hedefi:** React + Vite SPA iskeleti, API istemci katmanı ve tasarım sistemi temeli.
>
> **İlgili:** [`../tech-stack.md`](../tech-stack.md) · [`../architecture.md`](../architecture.md) §5 · [`../definition-of-done.md`](../definition-of-done.md)

## Sorumlu Rol(ler)
- Frontend (birincil), UX/UI (tasarım sistemi gözetimi)

## Bağımlılıklar
- [`01-architecture-tasks.md`](./01-architecture-tasks.md) Epik C (API uçları) — `health`/`version` tüketilebilir.

## Epikler ve Görevler

### Epik A: SPA iskeleti
- [x] A1 — Vite + React + TypeScript (strict) + **Tailwind CSS + Radix UI** proje kurulumu | Kabul: dev/build çalışır, lint temiz. Bkz. [`../tech-stack.md`](../tech-stack.md) §2.
- [x] A2 — Yönlendirme (router) ve sayfa düzeni iskeleti | Kabul: temel layout + 2 örnek rota.
- [x] A3 — Ortam/konfig yönetimi (API taban URL'i) | Kabul: build-time/run-time konfig stratejisi.

### Epik B: API istemci katmanı
- [x] B1 — Tipli HTTP istemci + hata yönetimi | Kabul: `health`/`version` çağrısı tipli döner.
- [x] B2 — Durum yönetimi entegrasyonu: **TanStack Query** (sunucu durumu) + **Zustand** (UI durumu) | Kabul: örnek sorgu/cache + global UI durumu çalışır.
- [x] B3 — Gerçek zamanlı olay (WS/SSE) istemci iskeleti | Kabul: bağlantı kurulup olay alınır.

### Epik C: Tasarım sistemi temeli
- [x] C1 — Token tabanlı tasarım sistemi (renk, tipografi, aralık) | Kabul: tüm stiller token üzerinden.
- [x] C1b — **Dinamik tema altyapısı (theme provider):** çalışma zamanında değiştirilebilir, token seti bazlı çoklu tema; özgün tema setleri **Cinematic (varsayılan)**, Aurora, Noir, Minimal (anahtarlar: `cinematic`/`aurora`/`noir`/`minimal`); seçimin kalıcılığı (oturumsuz: `localStorage`; oturumlu: kullanıcı profili — Sprint 03) | Kabul: tema anında değişir ve yeniden yüklemede hatırlanır. Yeni tema **yalnızca token seti eklemekle** gelebilmeli (kod değişikliği gerektirmez).
- [x] C2 — Temel bileşenler (buton, input, kart, modal, toast) | Kabul: erişilebilir, dokümante.
- [x] C3 — i18n altyapısı: kütüphane kurulumu, namespace/anahtar yapısı, en-US (kaynak) + tr-TR çeviri dosyaları, dil değiştirici, `Intl` ile tarih/sayı formatlama, tarayıcı dil tespiti | Kabul: dil değişimi çalışır, hardcoded metin yok, eksik anahtar geliştirme zamanı uyarısı verir. Bkz. [`../i18n-localization.md`](../i18n-localization.md).

## Kabul Kriterleri (Sprint Çıktısı)
- SPA derlenir, backend `health`/`version` bilgisini gösterir.
- Tasarım sistemi temel bileşenleri kullanıma hazır.

## Riskler
- Tasarım sistemi kararları tüm UI'yi etkiler → erken tutarlılık önemli.

## Kapsam Dışı
- Gerçek özellik ekranları (kütüphane, oynatıcı) — Sprint 05.
