# Sprint 10 — DevOps & Yayın Görevleri

> **Sprint hedefi:** Üretim kalitesinde sürüm yayını: imaj yayını, sürümleme, güncelleme yolu
> ve açık kaynak dağıtımı (1.0).
>
> **İlgili:** Sprint 01 [`../sprint-01-foundation/05-devops-tasks.md`](../sprint-01-foundation/05-devops-tasks.md) · [`../roadmap.md`](../roadmap.md) (M6)

## Sorumlu Rol(ler)
- DevOps/Platform (birincil), Tech Lead

## Bağımlılıklar
- Tüm sprintler; CI/CD (Sprint 01).

## Epikler ve Görevler

### Epik A: Sürüm yayını
- [ ] A1 — Otomatik sürümleme (SemVer) + changelog üretimi | Kabul: etiketli sürüm yayınlanır.
- [ ] A2 — Çok mimarili imaj yayını (registry) + `latest`/sürüm etiketleri | Kabul: kullanıcılar çekebilir.

### Epik B: Güncelleme & kalıcılık
- [ ] B1 — Güncelleme/migrasyon yolu (eski sürümden yeniye) doğrulaması | Kabul: veri kaybı olmadan yükseltme.
- [ ] B2 — Yedekleme/geri yükleme dokümante akışı | Kabul: kullanıcı veriyi koruyabilir.

### Epik C: Dağıtım kolaylığı
- [ ] C1 — Tek satır kurulum (`docker run`) + `docker compose` üretim örneği | Kabul: <10 dk kurulum (brief başarı kriteri).
- [ ] C2 — Sürüm sağlık/teşhis çıktısı | Kabul: destek için teşhis kolay.

## Kabul Kriterleri (Sprint Çıktısı)
- 1.0 imajı yayında; kurulum/güncelleme/yedekleme akışları doğrulanmış ve belgeli.

## Riskler
- Güncelleme/migrasyon hataları veri kaybı riski → sıkı test.

## Kapsam Dışı
- Bulut tek-tık dağıtım (managed) — gelecekte.
