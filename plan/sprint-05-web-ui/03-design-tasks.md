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
- [ ] A1 — somra logo, renk paleti, tipografi | Kabul: marka kılavuzu (mini).
- [ ] A2 — İkonografi ve görsel dil | Kabul: tutarlı set.

### Epik B: Akış tasarımları
- [ ] B1 — Ana sayfa, kütüphane, detay, oynatıcı ekran tasarımları | Kabul: onaylı mockup.
- [ ] B2 — Arama/filtre ve boş/hata durumları | Kabul: tüm durumlar tasarlanmış.

### Epik C: Tasarım sistemi olgunlaştırma
- [ ] C1 — Bileşen kütüphanesini genişletme (raf, kart, oynatıcı kontrolleri) | Kabul: frontend ile hizalı.
- [ ] C2 — Erişilebilirlik (kontrast, klavye, ARIA) kılavuzu | Kabul: WCAG temel uyum.

### Epik D: Dinamik tema tasarımı
- [ ] D1 — Dört özgün tema token setinin tasarımı (renk/tipografi/yoğunluk/kart-raf stili): **Cinematic (varsayılan)** — koyu, sinematik, sıcak vurgulu; **Aurora** — mavi/teal, ferah; **Noir** — derin mor/siyah, yüksek kontrast; **Minimal** — sade, nötr | Kabul: her tema için onaylı token paleti, tüm temalar WCAG kontrast eşiğini geçer, hiçbir tema bir markanın logosunu/görsel kimliğini taklit etmez.
- [ ] D2 — Tema seçici (theme switcher) UX'i ve önizleme | Kabul: tema değişimi anında ve tutarlı.
- [ ] D3 — Temaların tüm ekranlarda (ana sayfa, kütüphane, detay, oynatıcı) tutarlılık kontrolü | Kabul: hiçbir ekran tema-bağımsız kırılmaz.

## Kabul Kriterleri (Sprint Çıktısı)
- Onaylı görsel kimlik + temel ekran tasarımları + olgun tasarım sistemi + dört tema seti.

## Riskler
- Tasarım/uygulama uyumsuzluğu → frontend ile sıkı senkron.
- **Marka/legal:** Tema adları bilinçli olarak **özgün ve jeneriktir** (Cinematic/Aurora/Noir/Minimal); üçüncü taraf marka adı kullanılmaz. Tasarım kuralı: hiçbir tema bir ticari servisin logosunu, marka rengini birebir veya görsel kimliğini taklit etmez. Sprint 10 marka/lisans görevinde son kontrol yapılır. Bkz. [`../sprint-10-polish-oss-release/02-docs-tasks.md`](../sprint-10-polish-oss-release/02-docs-tasks.md).

## Kapsam Dışı
- Mobil/TV özel tasarımları — bu plan kapsamı dışında.
- Kullanıcının tamamen özel tema oluşturması (custom theme builder) — gelecekte.
