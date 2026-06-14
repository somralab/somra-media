# Sprint 03 — Veritabanı Görevleri (Kullanıcı Şeması)

> **Sprint hedefi:** Kullanıcı, rol, oturum, profil, ebeveyn kontrolü ve izleme durumu şeması.
>
> **İlgili:** [`../architecture.md`](../architecture.md) §4 · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (gözden geçirme)

## Bağımlılıklar
- Sprint 01 migrasyon altyapısı, Sprint 02 medya şeması.

## Epikler ve Görevler

### Epik A: Kimlik şeması
- [x] A1 — `user`, `role`, `permission`, `user_role` tabloları | Kabul: migrasyon + bütünlük.
- [x] A2 — `session`/token tablosu (çoklu cihaz) | Kabul: süre/iptal alanları.

### Epik B: Profil ve kontrol
- [x] B1 — `user_profile` (tercih, **dil/locale**, **arayüz teması**, avatar), ebeveyn kontrolü alanları (rating sınırı) | Kabul: çocuk profili kısıtları sorgulanabilir; dil ve tema tercihi saklanır (tema varsayılanı `cinematic`). Bkz. [`../i18n-localization.md`](../i18n-localization.md).

### Epik C: İzleme durumu
- [x] C1 — `watch_state` (ilerleme, izlendi), `favorite`, `watchlist` tabloları | Kabul: kullanıcı+öğe bazlı indeksler.

## Kabul Kriterleri (Sprint Çıktısı)
- Şema migrasyonlarla kurulur; backend akışları tutarlı çalışır.

## Riskler
- Yetki modeli sonraki sprintlerde genişler → esnek tasarım.

## Kapsam Dışı
- İstek yönetimi tabloları — Sprint 08.
