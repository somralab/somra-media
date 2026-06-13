# Somra — Proje Brifingi (Project Brief)

> Bu dosya projenin **tek doğruluk kaynağıdır (single source of truth)**. Tüm sprint ve görev
> dosyaları kapsam, hedef ve karar konularında bu dokümana referans verir. Bir görev bu
> dokümandaki kapsamla çelişiyorsa görev değil, bu doküman güncellenir (değişiklik kaydı ile).

İlgili dokümanlar: [`ideal-team.md`](./ideal-team.md) · [`architecture.md`](./architecture.md) · [`tech-stack.md`](./tech-stack.md) · [`roadmap.md`](./roadmap.md) · [`definition-of-done.md`](./definition-of-done.md) · [`i18n-localization.md`](./i18n-localization.md)

---

## 1. Vizyon

**Somra**, bir ev tipi medya sunucusunun ihtiyaç duyduğu tüm yetenekleri **tek bir çatı
altında**, **tek Docker kurulumuyla**, **minimum konfigürasyon / maksimum optimizasyon**
felsefesiyle sunan bütünleşik bir platformdur.

Bugün bir kullanıcı; Jellyfin/Emby/Plex (medya sunucu + transcode), Sonarr/Radarr/Lidarr
(içerik yönetimi), Prowlarr (indexer), Bazarr (altyazı), Overseerr/Jellyseerr (istek
yönetimi) ve indirme istemcileri gibi onlarca ayrı yazılımı kurmak, birbirine entegre etmek
ve tek tek ayarlamak zorundadır. Somra bu parçalı deneyimi **tek üründe** birleştirir.

## 2. Problem

- Çok sayıda ayrı yazılımın kurulumu, güncellenmesi ve entegrasyonu uzmanlık gerektirir.
- Her aracın kendi ayar/kimlik/veritabanı katmanı vardır; tutarsızlık ve bakım yükü oluşur.
- Optimizasyon (transcode profilleri, donanım hızlandırma, kalite profilleri) deneyim ister.
- Yeni kullanıcı için giriş bariyeri yüksektir.

## 3. Çözüm

Tek binary/servis odaklı, Go ile yazılmış bir çekirdek + React tabanlı tek arayüz. Kullanıcı
Docker ile kurar, bir kurulum sihirbazından geçer ve sistem **akıllı varsayılanlarla**
çalışmaya başlar. İleri kullanıcı her şeyi arayüzden ince ayar yapabilir.

## 4. Konsolide Kararlar

| Konu | Karar | Detay |
|---|---|---|
| Çekirdek strateji | **Sıfırdan kendi motorumuz** | Tarama, metadata, transcode, streaming bizim kodumuz. Bkz. [`architecture.md`](./architecture.md). |
| Backend | **Go** | Tek binary, düşük kaynak, yüksek eşzamanlılık. |
| Frontend | **React SPA (Vite)** | Web öncelikli arayüz. |
| Veritabanı | **Gömülü SQLite** | Sıfır konfig, tek dosya. |
| Transcode | **Önce yazılım (ffmpeg CPU)** | Donanım hızlandırma Sprint 07. |
| Platform | **Web öncelikli** | Mobil/TV gelecekteki yol haritasında, bu plan kapsamı dışında. |
| Kimlik | **Çoklu kullanıcı + RBAC + ebeveyn kontrolü** | Bkz. Sprint 03. |
| Çoklu dil (i18n) | **Çapraz kesen zorunluluk** | Kaynak dil **en-US**, çeviri **tr-TR**. Tam kapsam (UI, backend mesajları, bildirim, metadata dili, l10n, doküman). Bkz. [`i18n-localization.md`](./i18n-localization.md). |
| İçerik edinme (*arr/indexer/torrent/usenet) | **İlk fazda hariç, ileri sprintte tam** | Sprint 09'da eklenti mimarisiyle tam *arr/Prowlarr otomasyonu. |
| Ekip | **İdeal ekip varsayımı** | Bkz. [`ideal-team.md`](./ideal-team.md). |
| Takvim | **Katı deadline yok** | Sprint kadansı varsayılan 2 hafta; sprintler iş paketine göre boyutlandırılmıştır. |
| Marka | **somra** | Görsel kimlik Sprint 05 tasarım görevlerinde. |
| Lisans | **AGPL-3.0 + DCO (karar verildi)** | Aşağıya bakınız. |

## 5. Lisans: AGPL-3.0 + DCO (Karar Verildi)

**Karar: GNU Affero General Public License v3.0 (AGPL-3.0)** + katkı uzlaşısı olarak
**DCO (Developer Certificate of Origin)**. CLA kullanılmaz.

Gerekçe:
- Somra bir **sunucu/ağ servisi** yazılımıdır. MIT/Apache gibi permisif lisanslar, üçüncü
  tarafların kodu alıp **kapalı bir SaaS** olarak sunmasına ve değişiklikleri geri
  paylaşmamasına izin verir.
- AGPL-3.0, ağ üzerinden sunulan değişikliklerde bile **kaynak paylaşımını zorunlu** kılar.
  Bu, topluluk katkısını ve projenin açık kalmasını korur (Jellyfin'in GPL felsefesinin daha
  güçlü hâli).
- **DCO**, CLA'nın bürokrasisi olmadan katkı menşeini güvence altına alır (her commit `Signed-off-by`).
  Açık kaynak topluluğu için düşük sürtünmelidir.

Değerlendirilen ve elenen alternatifler:
- **Apache-2.0 / MIT**: SaaS sömürüsüne açık olduğu için elendi.
- **GPL-3.0**: Ağ maddesi (AGPL) tercih edildiği için elendi.
- **CLA**: Katkı sürtünmesini artırdığı için DCO lehine elendi.

> **Uygulama:** Repoda `LICENSE` (AGPL-3.0) ve `DCO`/`CONTRIBUTING` Sprint 01'de eklenir;
> resmî onay Sprint 01 mimari görevinde kaydedilir (yeniden tartışma değil, uygulama).

## 6. Kapsam (Bu Plan İçi)

Bu 10 sprintlik plan aşağıdaki yetenekleri kapsar:

1. Kütüphane tarama, teknik + zenginleştirilmiş metadata, dosya izleme.
2. Çoklu kullanıcı, RBAC, profiller, ebeveyn kontrolü, izleme durumu/devam.
3. Direct play + yazılım transcode, HLS/DASH, adaptif bitrate, altyazı.
4. Web arayüz: kütüphane gezinme, detay sayfaları, web oynatıcı, arama; **kullanıcı-seçimli dinamik tema** (özgün tema setleri: Cinematic varsayılan, Aurora, Noir, Minimal — kullanıcı bazında hatırlanır, marka taklidi yok).
5. Kurulum sihirbazı + akıllı varsayılanlar (minimum konfig / maksimum optimizasyon).
6. Donanım hızlandırma (QSV/NVENC/VAAPI/AMF).
7. İstek yönetimi (Overseerr işlevi) + bildirimler.
8. İndirme otomasyonu + indexer (torrent + usenet) + eklenti mimarisi (*arr/Prowlarr işlevi).
9. Açık kaynak yayın hazırlığı.
10. **Çoklu dil (i18n/l10n):** en-US (kaynak) + tr-TR; geliştirme süreci ilk günden i18n-uyumlu. Detay: [`i18n-localization.md`](./i18n-localization.md).

## 7. Kapsam Dışı (Bu Plan Dışı — Drift Önleme)

Aşağıdakiler **bilinçli olarak bu planın dışındadır**. Bir görev bunlardan birine girerse,
önce bu doküman güncellenmeli ve yeni sprint/efor planlanmalıdır:

- Native mobil ve TV uygulamaları (iOS/Android/Android TV/tvOS/webOS/Tizen).
- Canlı TV / DVR / EPG.
- Çoklu sunucu federasyonu / bulut senkron / harici CDN.
- Müzik için gelişmiş özellikler (tam Lidarr paritesi) — temel müzik kütüphanesi dışında.
- DLNA/Chromecast tam uyumluluğu (sadece web öncelikli).

## 8. Başarı Kriterleri (Üst Düzey)

- **Kurulum:** Sıfır bilgi seviyesinde bir kullanıcı, tek `docker run`/`docker compose up`
  ile çalışan bir sunucuya 10 dakikadan kısa sürede ulaşır.
- **Optimizasyon:** Donanım/medya tespitine göre otomatik transcode profili; kullanıcı manuel
  ayar yapmadan akıcı oynatma alır.
- **Bütünlük:** Yukarıdaki ayrı yazılımların temel işlevleri tek arayüzden, tek kimlik ve tek
  veri katmanı üzerinden kullanılabilir.
- **Performans:** Ev sunucusu donanımında (ör. 4 çekirdek/8GB) makul kaynakla çalışır.
- **Açık kaynak hazırlığı:** Lisans, katkı rehberi, güvenlik politikası, CI/CD ve dokümantasyon hazır.

## 9. Yönetişim Kuralları (Anti-Drift / Çerçeveleme)

> Bu kurallar tüm sprint ve görev dosyaları için **bağlayıcıdır**.

1. **Kapsam kilidi:** Hiçbir sprint, [Bölüm 7](#7-kapsam-d%C4%B1%C5%9F%C4%B1-bu-plan-d%C4%B1%C5%9F%C4%B1--drift-%C3%B6nleme)'deki kapsam dışı maddelere bu doküman güncellenmeden giremez.
2. **DoD zorunlu:** Her görev [`definition-of-done.md`](./definition-of-done.md) ölçütlerini karşılamadan "bitti" sayılmaz.
3. **Bağımlılık disiplini:** Bir sprint, önceki sprintlerin tamamlanan çıktılarına dayanır; bağımlılıklar görev dosyalarında açıkça belirtilir.
4. **Tek doğruluk kaynağı:** Kararlar burada toplanır; çelişki halinde bu doküman esas alınır.
5. **Sürüm etiketleme:** Her sprint sonunda çalışabilir bir artımlı sürüm (milestone) çıkarılır. Bkz. [`roadmap.md`](./roadmap.md).
6. **i18n zorunluluğu:** Kullanıcıya görünen hiçbir metin koda gömülemez; her özellik en-US + tr-TR ile birlikte gelir. Bkz. [`i18n-localization.md`](./i18n-localization.md) ve [`definition-of-done.md`](./definition-of-done.md).

## 10. Değişiklik Kaydı

| Tarih | Değişiklik | Sorumlu |
|---|---|---|
| (oluşturma) | İlk brifing oluşturuldu | — |
