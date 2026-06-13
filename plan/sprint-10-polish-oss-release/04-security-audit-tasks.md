# Sprint 10 — Güvenlik Denetimi Görevleri

> **Sprint hedefi:** Açık kaynak yayını öncesi kapsamlı güvenlik denetimi ve güvenli
> varsayılanların doğrulanması.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) §6 · Sprint 03 [`../sprint-03-auth-users/04-security-tasks.md`](../sprint-03-auth-users/04-security-tasks.md)

## Sorumlu Rol(ler)
- Tech Lead (birincil), Backend, QA

## Bağımlılıklar
- Tüm sprintlerin güvenlikle ilgili çıktıları.

## Epikler ve Görevler

### Epik A: Güvenlik denetimi
- [ ] A1 — Kimlik/yetki/oturum yüzeyi denetimi (RBAC kapsam doğrulaması) | Kabul: yetki ihlali bulunmaz.
- [ ] A2 — Girdi doğrulama/enjeksiyon/SSRF/path traversal denetimi | Kabul: bilinen sınıflar kapatılır.
- [ ] A3 — Bağımlılık güvenlik taraması (SCA) + lisans uyumu | Kabul: kritik açık/uyumsuz lisans yok.

### Epik B: Güvenli varsayılanlar
- [ ] B1 — Varsayılan yapılandırma sertleştirme (sırlar, CORS, rate limit, HTTPS rehberi) | Kabul: güvenli varsayılan.
- [ ] B2 — Eklenti izolasyonu güvenlik gözden geçirmesi | Kabul: eklenti çekirdeği tehlikeye atamaz.

### Epik C: Süreç
- [ ] C1 — Güvenlik politikası (SECURITY.md) + açık bildirim süreci | Kabul: yayında mevcut.

## Kabul Kriterleri (Sprint Çıktısı)
- Kritik/yüksek güvenlik bulgusu kalmamış; güvenli varsayılanlar ve güvenlik politikası hazır.

## Riskler
- Geç bulunan açık yayını geciktirir → denetim sprint başında başlamalı.

## Kapsam Dışı
- Resmi 3. taraf penetrasyon testi sertifikası — opsiyonel/gelecekte.
