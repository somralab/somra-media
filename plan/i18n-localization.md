# Somra — Çoklu Dil & Yerelleştirme (i18n / l10n)

> Çoklu dil, projenin **çapraz kesen (cross-cutting) ve bağlayıcı** bir gereksinimidir.
> Geliştirme sürecinin **ilk günden** i18n-uyumlu yürütülmesi zorunludur: hiçbir kullanıcıya
> görünen metin koda gömülü (hardcoded) yazılamaz. Bu kural [`definition-of-done.md`](./definition-of-done.md)
> ile denetlenir.

İlgili: [`project-brief.md`](./project-brief.md) · [`architecture.md`](./architecture.md) · [`tech-stack.md`](./tech-stack.md) · [`definition-of-done.md`](./definition-of-done.md) · [`roadmap.md`](./roadmap.md)

---

## 1. Hedef Diller

| Dil | Kod | Rol |
|---|---|---|
| İngilizce (ABD) | **en-US** | **Kaynak dil + yedek (fallback).** Tüm anahtarların referansı. |
| Türkçe (Türkiye) | **tr-TR** | Birinci ek çeviri; ürün önceliği. |

- **Kaynak dil en-US'tir:** Tüm metin anahtarları önce en-US'te tanımlanır; çeviri eksikse en-US'e düşülür.
- Yeni dil eklemek **yalnızca çeviri dosyası eklemekle** mümkün olacak şekilde tasarlanır (kod değişikliği gerektirmez).

## 2. Kapsam

i18n/l10n aşağıdaki tüm katmanları kapsar:

1. **Frontend arayüz metinleri** (zorunlu) — tüm etiket, buton, mesaj, boş/hata durumları.
2. **Backend mesajları** — API hata/doğrulama mesajları (anahtar + dil pazarlığı ile).
3. **Bildirim şablonları** — e-posta / Discord / webhook (TR + EN şablon). Bkz. Sprint 08.
4. **Medya metadata dili** — film/dizi açıklama/başlık, kullanıcı diline göre sağlayıcıdan çekilir. Bkz. Sprint 02.
5. **Tarih / sayı / format yerelleştirme (l10n)** — yerel ayara uygun gösterim.
6. **Dokümantasyon** — kullanıcı ve geliştirici dokümanı TR + EN. Bkz. Sprint 10.

## 3. Dil Seçimi (Negotiation)

Öncelik sırası (yüksekten düşüğe):

1. **Kullanıcı profili tercihi** (oturum açmış kullanıcı) — kalıcı.
2. **Sistem varsayılanı** (admin tarafından kurulum/ayarda belirlenir).
3. **Tarayıcı otomatik tespiti** (`Accept-Language`) — oturumsuz/ilk ziyaret.
4. **Yedek:** en-US.

- Kullanıcı her zaman sistem varsayılanını **override** edebilir (per-user).
- Backend, isteklerde dil bağlamını (locale) tüm mesaj üretimine taşır.

## 4. Teknik Yaklaşım

### 4.1 Çeviri dosyaları (repo içi)
- Çeviriler **repoda** tutulur (JSON), PR ile katkı alınır.
- Namespace/anahtar yapısı modüllere göre düzenlenir (`common`, `library`, `player`, `auth`, `errors`).
- Anahtar adlandırma standardı: `domain.context.key`.
- Topluluk çevirisi için **Weblate** (self-host, git entegre) kullanılır. Kurulum Sprint 10'da.

### 4.2 Frontend
- i18n kütüphanesi: **`i18next` + `react-i18next`** (bkz. [`tech-stack.md`](./tech-stack.md) §2).
- Çoğul (pluralization), enterpolasyon, tarih/sayı formatlama (`Intl`) desteği.
- Eksik anahtar tespiti için geliştirme zamanı uyarısı + CI kontrolü.

### 4.3 Backend
- Mesaj kataloğu: **`nicksnyder/go-i18n/v2` + `golang.org/x/text`** (locale eşleştirme).
- API hatalarında hem makine-okur kod hem yerelleştirilmiş mesaj döner.

## 5. Bağlayıcı Geliştirme Kuralları (Anti-Drift)

> Bu kurallar tüm sprint/görev dosyaları için zorunludur ve [`definition-of-done.md`](./definition-of-done.md)'de denetlenir.

1. **Hardcoded metin yasağı:** Kullanıcıya görünen hiçbir metin doğrudan koda yazılamaz; daima anahtar üzerinden çözülür.
2. **Kaynak dil bütünlüğü:** Her yeni özellik en-US anahtarlarıyla **birlikte** gelir; eksik anahtar CI'da hata verir.
3. **tr-TR eşzamanlı:** Özellik tamamlandığında tr-TR çevirisi de eklenir (eksikse görev "Done" sayılmaz).
4. **Format güvenliği:** Tarih/sayı/para gösterimi yerel ayar API'leriyle yapılır; elle string birleştirme yok.
5. **Yön/uzunluk:** UI, metin uzunluğu değişimlerine (TR↔EN) dayanıklı tasarlanır.

## 6. Test & Kalite

- **CI kontrolü:** Eksik/fazla anahtar, kullanılmayan anahtar tespiti.
- **QA:** Her sprintte yeni metinlerin iki dilde de doğruluğu; pseudo-locale ile taşma/uzunluk testi.
- **Yayın kapısı (Sprint 10):** en-US ve tr-TR için **%100 anahtar tamlığı** kabul kriteridir.

## 7. Sprint Dağılımı (Çapraz Kesen)

| Sprint | i18n katkısı |
|---|---|
| 01 | Frontend + backend i18n altyapısı, anahtar standardı, CI kontrolü, dil pazarlığı iskeleti |
| 02 | Metadata dilinin kullanıcı locale'ine göre çekilmesi |
| 03 | Kullanıcı profili dil tercihi + tarayıcı otomatik tespiti |
| 06 | Onboarding'de dil seçimi + sistem varsayılan dil ayarı |
| 08 | Bildirim şablonlarının TR/EN yerelleştirmesi |
| 10 | Doküman TR+EN, çeviri tamlık QA'i, (opsiyonel) çeviri platformu kararı |

## 8. Kararlar (Kapatıldı)
| Karar | Sonuç |
|---|---|
| Frontend i18n kütüphanesi | `i18next` + `react-i18next` |
| Backend i18n kütüphanesi | `nicksnyder/go-i18n/v2` + `golang.org/x/text` |
| Çeviri dosya formatı | JSON, repo içi |
| Harici çeviri platformu | **Weblate** (self-host, OSS, git entegre) — kurulum Sprint 10 |
