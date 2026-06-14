# Sprint 08 — Bildirim Görevleri

> **Sprint hedefi:** Olay tabanlı bildirim altyapısı (webhook, Discord, e-posta) ve istek/sistem
> olaylarıyla entegrasyon.
>
> **İlgili:** [`../tech-stack.md`](../tech-stack.md) §5 · [`01-backend-tasks.md`](./01-backend-tasks.md)

## Sorumlu Rol(ler)
- Backend (birincil)

## Bağımlılıklar
- [`01-backend-tasks.md`](./01-backend-tasks.md) (istek olayları), scheduler/olay altyapısı (Sprint 01).

## Epikler ve Görevler

### Epik A: Bildirim altyapısı
- [x] A1 — Olay → bildirim soyutlaması (kanal eklenebilir) | Kabul: yeni kanal kolayca eklenir.
- [x] A2 — Şablon ve dil desteği (i18n): alıcının dil tercihine göre tr-TR/en-US şablon seçimi, eksikse en-US yedeği | Kabul: bildirim alıcının dilinde gönderilir. Bkz. [`../i18n-localization.md`](../i18n-localization.md) §2.

### Epik B: Kanal entegrasyonları
- [x] B1 — Webhook (genel) | Kabul: yapılandırılabilir webhook tetiklenir.
- [x] B2 — Discord + e-posta (SMTP) | Kabul: test gönderimi başarılı.

### Epik C: Olay bağlama
- [x] C1 — İstek olayları (oluştu/onaylandı/tamamlandı) ve sistem olayları (hata) | Kabul: doğru olayda bildirim.
- [x] C2 — Kullanıcı/admin bildirim tercihleri | Kabul: abonelik yönetimi.

## Kabul Kriterleri (Sprint Çıktısı)
- Olaylar yapılandırılabilir kanallardan bildirilir; tercihler yönetilir.

## Riskler
- Spam/aşırı bildirim → tercih + debounce.

## Kapsam Dışı
- Push/native mobil bildirim — bu plan kapsamı dışında.
