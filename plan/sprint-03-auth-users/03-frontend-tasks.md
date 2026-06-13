# Sprint 03 — Frontend Görevleri (Kimlik & Kullanıcı UI)

> **Sprint hedefi:** Giriş/çıkış, kullanıcı yönetimi, profil ve ebeveyn kontrolü arayüzleri.
>
> **İlgili:** Sprint 01 SPA iskeleti · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Sorumlu Rol(ler)
- Frontend (birincil)

## Bağımlılıklar
- Bu sprint backend kimlik/RBAC API'leri.

## Epikler ve Görevler

### Epik A: Kimlik akışları
- [ ] A1 — Giriş/çıkış ekranları + oturum yönetimi (token saklama, yenileme) | Kabul: güvenli, korumalı rotalar.
- [ ] A2 — Korumalı rota / yetki bazlı UI gösterimi | Kabul: yetkisiz kullanıcı kısıtlı görür.

### Epik B: Kullanıcı yönetimi (admin)
- [ ] B1 — Kullanıcı listesi/oluştur/düzenle + rol atama | Kabul: RBAC API'sine bağlı.
- [ ] B2 — Ebeveyn kontrolü ayar arayüzü | Kabul: rating sınırı ayarlanabilir.

### Epik C: Profil
- [ ] C1 — Profil düzenleme (dil seçimi tr-TR/en-US, **tema seçimi** Cinematic/Aurora/Noir/Minimal, avatar, tercihler) | Kabul: dil ve tema değişimi anında arayüze uygulanır ve kalıcı saklanır; tarayıcı dili ile otomatik ön seçim; oturumsuzken `localStorage`'taki tema oturum açılınca profile taşınır. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §3 ve [`04-frontend-tasks.md`](../sprint-01-foundation/04-frontend-tasks.md) C1b.

## Kabul Kriterleri (Sprint Çıktısı)
- Kullanıcı giriş yapar, admin kullanıcıları yönetir, ebeveyn kontrolü ayarlanır.

## Riskler
- Token saklama güvenliği → güvenli pratikler (Sprint 03 güvenlik görevleriyle uyumlu).

## Kapsam Dışı
- Oynatıcı/gezinme ekranları — Sprint 05.
