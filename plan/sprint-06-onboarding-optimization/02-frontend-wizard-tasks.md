# Sprint 06 — Frontend Görevleri (Kurulum Sihirbazı & Ayarlar)

> **Sprint hedefi:** İlk kurulum sihirbazı (minimum adım, maksimum otomatik) ve ayar arayüzleri.
>
> **İlgili:** [`01-backend-tasks.md`](./01-backend-tasks.md) · [`../project-brief.md`](../project-brief.md) (kurulum başarı kriteri: <10 dk)

## Sorumlu Rol(ler)
- Frontend (birincil), UX/UI (akış)

## Bağımlılıklar
- [`01-backend-tasks.md`](./01-backend-tasks.md) (tespit + ayar + onboarding state).

## Epikler ve Görevler

### Epik A: Kurulum sihirbazı
- [x] A0 — İlk adım: dil seçimi (tr-TR/en-US, tarayıcı diliyle ön seçili) + sistem varsayılan dili belirleme | Kabul: seçilen dil tüm sihirbaza anında uygulanır. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §3.
- [x] A1 — Adım akışı: admin oluştur → kütüphane ekle → otomatik tespit/öneri onayı → bitir | Kabul: <10 dk'da çalışır sistem.
- [x] A2 — Akıllı varsayılanları gösterip "öner ve uygula" deneyimi | Kabul: kullanıcı manuel ayar yapmadan ilerler.
- [x] A3 — İlk tarama ilerlemesini sihirbaz içinde gösterme | Kabul: canlı geri bildirim.

### Epik B: Ayar arayüzü
- [x] B1 — Kategori bazlı ayar ekranları (genel, kütüphane, oynatma, kullanıcılar) | Kabul: backend ayar API'sine bağlı.
- [x] B2 — "Gelişmiş" gizli/açık modu (basit varsayılan, isteyene detay) | Kabul: minimum konfig felsefesi.

## Kabul Kriterleri (Sprint Çıktısı)
- Yeni kullanıcı sihirbazdan geçip çalışan, optimize bir sunucuya hızla ulaşır.

## Riskler
- Sihirbaz çok uzun olursa felsefeye aykırı → adım sayısı minimumda tutulur.

## Kapsam Dışı
- GPU seçim arayüzü — Sprint 07.
