# Sprint 03 — Backend Görevleri (Kimlik & RBAC)

> **Sprint hedefi:** Çoklu kullanıcı, kimlik doğrulama, RBAC, profiller, ebeveyn kontrolü ve
> izleme durumu (resume) backend'i.
>
> **İlgili:** [`../architecture.md`](../architecture.md) §3/§5 · [`../project-brief.md`](../project-brief.md) (Kimlik kararı) · [`04-security-tasks.md`](./04-security-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (oturum stratejisi)

## Bağımlılıklar
- Sprint 01 (API/oturum kararı, veri katmanı), Sprint 02 (medya öğeleri — izleme durumu için).

## Epikler ve Görevler

### Epik A: Kimlik doğrulama
- [x] A1 — Kullanıcı kaydı/giriş, parola hash (güçlü algoritma), oturum/token yönetimi | Kabul: güvenli akış, test edilir.
- [x] A2 — Oturum yenileme/çıkış, çoklu cihaz oturumu | Kabul: oturumlar yönetilebilir.
- [x] A3 — İlk kurulumda admin oluşturma akışı | Kabul: Sprint 06 onboarding ile uyumlu.

### Epik B: RBAC ve profiller
- [x] B1 — Rol/yetki modeli (admin, kullanıcı, çocuk) + yetki kontrol middleware'i | Kabul: korumalı uçlar yetkiye göre filtreler.
- [x] B2 — Kullanıcı profilleri (avatar, dil tercihi `tr-TR`/`en-US`, **arayüz teması** `cinematic`/`aurora`/`noir`/`minimal`, tercihler) | Kabul: profil CRUD; dil tercihi dil pazarlığında en yüksek önceliklidir; tema tercihi kalıcı saklanır (varsayılan `cinematic`). Bkz. [`../i18n-localization.md`](../i18n-localization.md) §3.
- [x] B3 — Ebeveyn kontrolü: yaş derecesi (rating) sınırı, içerik kısıtı | Kabul: kısıtlı içerik çocuk profilinde gizlenir.

### Epik C: İzleme durumu
- [x] C1 — İzleme ilerleme/devam (resume) ve "izlendi" durumu | Kabul: kaldığı yerden devam.
- [x] C2 — Kullanıcı bazlı favoriler/izleme listesi | Kabul: CRUD + filtre.

## Kabul Kriterleri (Sprint Çıktısı)
- Çoklu kullanıcı giriş yapar; roller ve ebeveyn kısıtları uygulanır; izleme durumu tutulur.
- [`04-security-tasks.md`](./04-security-tasks.md) gereksinimleri karşılanır.

## Riskler
- Güvenlik kritik → Sprint 10 güvenlik denetimiyle de doğrulanır.

## Kapsam Dışı
- Harici kimlik (OIDC/LDAP) — bu plan kapsamı dışında (gelecekte).
