# Sprint 03 — Güvenlik Görevleri

> **Sprint hedefi:** Kimlik ve yetkilendirme katmanının güvenlik temellerini kurmak.
>
> **İlgili:** [`../definition-of-done.md`](../definition-of-done.md) §6 · [`01-backend-tasks.md`](./01-backend-tasks.md) · Sprint 10 güvenlik denetimi

## Sorumlu Rol(ler)
- Backend (birincil), Tech Lead (gözetim)

## Bağımlılıklar
- [`01-backend-tasks.md`](./01-backend-tasks.md), [`02-database-tasks.md`](./02-database-tasks.md)

## Epikler ve Görevler

### Epik A: Kimlik güvenliği
- [ ] A1 — Güçlü parola hash + politika | Kabul: zayıf parola reddedilir.
- [ ] A2 — Brute-force koruması (rate limit, kilitleme) | Kabul: tekrarlı başarısız giriş sınırlanır.
- [ ] A3 — Güvenli oturum/token (süre, iptal, yenileme) | Kabul: çalınan token iptal edilebilir.

### Epik B: Yetki güvenliği
- [ ] B1 — Tüm korumalı uçlarda yetki kontrolü zorunluluğu | Kabul: yetkisiz erişim engellenir, test edilir.
- [ ] B2 — Girdi doğrulama ve enjeksiyon koruması | Kabul: parametreli sorgular, doğrulama.

### Epik C: Sır yönetimi
- [ ] C1 — Sağlayıcı anahtarları ve sırların güvenli saklanması | Kabul: koda gömülü sır yok.

## Kabul Kriterleri (Sprint Çıktısı)
- Kimlik/yetki katmanı güvenli varsayılanlarla çalışır; temel saldırı senaryolarına dayanıklı.

## Riskler
- Eksik yetki kontrolü kritik açık → kapsamlı test ve Sprint 10 denetimi.

## Kapsam Dışı
- Tam güvenlik denetimi/penetrasyon — Sprint 10.
