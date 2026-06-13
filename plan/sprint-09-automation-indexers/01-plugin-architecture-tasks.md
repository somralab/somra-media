# Sprint 09 — Eklenti Mimarisi Görevleri

> **Sprint hedefi:** İçerik edinme yeteneklerini (indexer + indirme istemcisi) **çekirdekten
> izole** eden eklenti çerçevesi. Yasal risk izolasyonu için kritik.
>
> **İlgili:** [`../architecture.md`](../architecture.md) §6 (eklenti) · [`../project-brief.md`](../project-brief.md) §7 (kapsam/legal) · [`02-indexer-integration-tasks.md`](./02-indexer-integration-tasks.md)

## Sorumlu Rol(ler)
- Tech Lead (birincil — mimari), Backend (uygulama)

## Bağımlılıklar
- Sprint 01 (mimari/modül sınırları), Sprint 08 (otomasyon handoff arayüzü).

## Epikler ve Görevler

### Epik A: Eklenti sözleşmesi
- [ ] A1 — `Indexer` ve `DownloadClient` arayüz sözleşmeleri (Search, Capabilities, Add, Status) | Kabul: net, sürümlü arayüz.
- [ ] A2 — Eklenti yaşam döngüsü (kayıt, etkinleştir, yapılandır, devre dışı) | Kabul: çalışma zamanında yönetilir.

### Epik B: İzolasyon & güvenlik
- [ ] B1 — Çekirdeğin eklentilerden bağımsız çalışması (eklenti yokken sistem tam fonksiyonel) | Kabul: eklentisiz mod test edilir.
- [ ] B2 — Eklenti yapılandırma/sır yönetimi (güvenli saklama) | Kabul: sırlar korunur.
- [ ] B3 — Eklenti dağıtım/paketleme stratejisi (ayrı paketlenebilir) | Kabul: legal izolasyon dokümante.

### Epik C: Yönetim API'si
- [ ] C1 — Eklenti listeleme/yapılandırma/test bağlantısı API'si | Kabul: frontend ile koordine.

## Kabul Kriterleri (Sprint Çıktısı)
- Çekirdek tarafsız; indexer/indirme yetenekleri takılabilir eklentiler olarak çalışır ve izoledir.

## Riskler
- **Yasal risk** → izolasyon ve net lisans şart (bkz. [`../project-brief.md`](../project-brief.md) §5/§7).

## Kapsam Dışı
- Üçüncü taraf eklenti pazar yeri — gelecekteki yol haritası.
