# Sprint 05 — Tasarım Görevleri (UX/UI)

> **Sprint hedefi:** Görsel kimlik (somra markası), gezinme/oynatma akış tasarımları ve
> tasarım sistemini olgunlaştırmak.
>
> **İlgili:** [`../project-brief.md`](../project-brief.md) (marka: somra) · Sprint 01 tasarım sistemi temeli · [`01-frontend-tasks.md`](./01-frontend-tasks.md)

## Sorumlu Rol(ler)
- UX/UI Tasarımcı (birincil), Frontend (uygulama)

## Bağımlılıklar
- Sprint 01 tasarım sistemi temeli.

## Epikler ve Görevler

### Epik A: Marka & görsel kimlik
- [x] A1 — somra logo, renk paleti, tipografi | Kabul: marka kılavuzu (mini).
- [x] A2 — İkonografi ve görsel dil | Kabul: tutarlı set.

### Epik B: Akış tasarımları
- [x] B1 — Ana sayfa, kütüphane, detay, oynatıcı ekran tasarımları | Kabul: onaylı mockup.
- [x] B2 — Arama/filtre ve boş/hata durumları | Kabul: tüm durumlar tasarlanmış.

### Epik C: Tasarım sistemi olgunlaştırma
- [x] C1 — Bileşen kütüphanesini genişletme (raf, kart, oynatıcı kontrolleri) | Kabul: frontend ile hizalı.
- [x] C2 — Erişilebilirlik (kontrast, klavye, ARIA) kılavuzu | Kabul: WCAG temel uyum.

### Epik D: Dinamik tema tasarımı
- [x] D1 — Dört özgün tema token setinin tasarımı
- [x] D2 — Tema seçici (theme switcher) UX'i ve önizleme
- [x] D3 — Temaların tüm ekranlarda tutarlılık kontrolü

## Kabul Kriterleri (Sprint Çıktısı)
- Onaylı görsel kimlik + temel ekran tasarımları + olgun tasarım sistemi + dört tema seti.

## Riskler
- Tasarım/uygulama uyumsuzluğu → frontend ile sıkı senkron.
- **Marka/legal:** Tema adları bilinçli olarak **özgün ve jeneriktir** (Cinematic/Aurora/Noir/Minimal); üçüncü taraf marka adı kullanılmaz. Tasarım kuralı: hiçbir tema bir ticari servisin logosunu, marka rengini birebir veya görsel kimliğini taklit etmez. Sprint 10 marka/lisans görevinde son kontrol yapılır. Bkz. [`../sprint-10-polish-oss-release/02-docs-tasks.md`](../sprint-10-polish-oss-release/02-docs-tasks.md).

## Kapsam Dışı
- Mobil/TV özel tasarımları — bu plan kapsamı dışında.
- Kullanıcının tamamen özel tema oluşturması (custom theme builder) — gelecekte.
