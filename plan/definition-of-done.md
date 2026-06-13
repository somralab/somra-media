# Somra — Definition of Done & Çalışma Kuralları

> Bu doküman **bağlayıcı kurallar** içerir. Bir görev/PR/sprint, buradaki ölçütleri
> karşılamadan "tamamlandı" sayılamaz. Tüm görev dosyaları bu dokümana referans verir.

İlgili: [`project-brief.md`](./project-brief.md) · [`tech-stack.md`](./tech-stack.md) · [`architecture.md`](./architecture.md) · [`i18n-localization.md`](./i18n-localization.md)

---

## 1. Görev Seviyesi DoD

Bir görev "Done" sayılması için:

1. Kabul kriterleri (görev dosyasındaki) karşılanmış ve doğrulanmıştır.
2. Kod, [`architecture.md`](./architecture.md) modül sınırlarına uyar.
3. Birim testleri yazılmış ve geçiyor; ilgili yerde entegrasyon testi var; coverage eşikleri (§4.1) karşılanır.
4. Lint/format temiz (Go: `golangci-lint`; FE: ESLint + Prettier).
5. Gerekli dokümantasyon/README güncellemesi yapılmış.
6. En az 1 kod inceleme onayı (mimari etkisi varsa Tech Lead onayı).
7. CI tüm aşamalarda yeşil.
8. **i18n uyumu:** Kullanıcıya görünen tüm metinler anahtar üzerinden çözülür (hardcoded yok); en-US **ve** tr-TR anahtarları eklenmiştir. Bkz. [`i18n-localization.md`](./i18n-localization.md) §5.

## 2. Sprint Seviyesi DoD

1. Sprint hedefindeki tüm "must" görevler Done.
2. Çalışan **artımlı sürüm** demo edilebilir.
3. Regresyon paketi geçiyor (QA).
4. Bilinen kritik/yüksek hata kalmamış.
5. Bağımlı sonraki sprintin başlayabilmesi için gereken çıktılar yayınlanmış.

## 3. Kod Standartları (Özet)

- **Go:** `gofmt`/`goimports`, anlamlı paket sınırları, hata sarmalama (`%w`), context kullanımı,
  global durumdan kaçınma. Genel API'ler dökümante edilir.
- **TypeScript/React:** strict mode, tipli API katmanı, bileşen/hook ayrımı, erişilebilirlik.
- **Yorumlar:** "ne yaptığını" anlatan gereksiz yorum yok; yalnızca niyet/kısıt/ödünleşim açıklanır.
- **Commit/PR:** küçük, odaklı; PR açıklamasında ilgili görev ve kabul kriteri referansı.

## 4. Test Politikası

| Katman | Beklenti |
|---|---|
| Birim | İş kuralları ve saf fonksiyonlar için zorunlu. |
| Entegrasyon | DB, dosya tarama, transcode pipeline gibi sınırlar için. |
| E2E | Kritik kullanıcı akışları (giriş, gezinme, oynatma, kurulum sihirbazı). |
| Performans | Streaming/transcode ve tarama için temel ölçümler (Sprint 04+). |

### 4.1 Coverage (Kapsam) Standardı — Bağlayıcı

| Alan | Minimum satır kapsamı |
|---|---|
| Çekirdek iş mantığı (Go) | **≥ %80** |
| Kritik modüller (Go): kimlik/RBAC, tarama, transcode karar motoru, otomasyon import | **≥ %90** |
| Frontend bileşen testleri | **≥ %70** |
| Frontend kritik akışlar (giriş, gezinme, oynatma, kurulum sihirbazı) | **e2e zorunlu** (yüzde değil, akış kapsamı) |

- Eşikler **CI'da ölçülür ve zorunludur**; eşik altındaki PR merge **edilemez** (bkz. §5).
- Kritik modül listesi mimari değiştikçe Tech Lead onayıyla güncellenir.
- Coverage bir hedeftir, anlamsız test yazımıyla şişirilmez; inceleme bunu denetler.

## 5. CI Kapıları (Gate)

PR merge için zorunlu yeşil aşamalar: `lint` → `i18n-check` → `unit-test` → `integration-test` → `coverage-gate` → `build` → `image-build`.
Bu kapılardan biri kırmızıysa merge **yapılmaz**.

> `i18n-check`: eksik/kullanılmayan çeviri anahtarı ve en-US/tr-TR tamlık kontrolü. Bkz. [`i18n-localization.md`](./i18n-localization.md) §6.
>
> `coverage-gate`: §4.1 eşiklerini ölçer (çekirdek ≥%80, kritik modüller ≥%90, frontend bileşen ≥%70) ve coverage raporu üretir; eşik altında merge engellenir.

## 6. Güvenli Varsayılanlar

- En az yetki ilkesi; her girdi doğrulanır.
- Sırlar koda gömülmez; ortam değişkeni/secret yönetimi.
- Harici sağlayıcı anahtarları kullanıcı tarafından girilir, güvenli saklanır.

## 7. Anti-Drift (Kapsam Koruma) Kuralları

> [`project-brief.md`](./project-brief.md) §9 yönetişim kurallarının operasyonel karşılığı.

1. Görev, ait olduğu sprintin hedefi dışına çıkamaz; çıkıyorsa yeni görev/sprint açılır.
2. [`project-brief.md`](./project-brief.md) §7 kapsam-dışı maddelere dokunan iş, brief güncellenmeden başlamaz.
3. "Yapılması iyi olur" türü ek işler backlog'a yazılır, sprint içinde sessizce eklenmez.
4. Her görev dosyası: **Hedef, Sorumlu rol(ler), Görevler, Bağımlılıklar, Kabul kriterleri, Riskler, Kapsam dışı** bölümlerini içerir.

## 8. Görev Dosyası Şablonu (Tüm sprintlerde kullanılır)

```md
# Sprint NN — <Disiplin> Görevleri

> Sprint hedefi: ...
> İlgili: project-brief.md, architecture.md, definition-of-done.md, (önceki sprintler)

## Sorumlu Rol(ler)
...

## Bağımlılıklar
...

## Epikler ve Görevler
### Epik A: ...
- [ ] Görev A1 — <açıklama> | Kabul: <ölçüt>
...

## Kabul Kriterleri (Sprint Çıktısı)
...

## Riskler
...

## Kapsam Dışı
...
```
