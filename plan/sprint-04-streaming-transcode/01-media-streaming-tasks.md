# Sprint 04 — Medya & Streaming Görevleri

> **Sprint hedefi:** Direct play + yazılım (CPU) transcode, HLS/DASH paketleme, adaptif
> bitrate (ABR) ve altyazı/ses kanalı işleme. M3 (ilk uçtan uca oynatma) çekirdeği.
>
> **İlgili:** [`../architecture.md`](../architecture.md) §3 (Streaming) · [`../project-brief.md`](../project-brief.md) (transcode kararı: önce yazılım) · [`../tech-stack.md`](../tech-stack.md)

## Sorumlu Rol(ler)
- Medya/Streaming Uzmanı (birincil), Backend (destek)

## Bağımlılıklar
- Sprint 02 (medya/teknik metadata), Sprint 03 (yetki — kim neyi oynatabilir).

## Epikler ve Görevler

### Epik A: Oynatma kararı
- [x] A1 — İstemci yetenek (capability) profili alımı (desteklenen kodek/konteyner) | Kabul: istemci profili ile eşleştirme.
- [x] A2 — Direct play / direct stream / transcode karar motoru | Kabul: gereksiz transcode yapılmaz.

### Epik B: Transcode pipeline (yazılım)
- [x] B1 — ffmpeg süreç yönetimi (başlat/izle/sonlandır, kaynak sınırı) | Kabul: artık (zombie) süreç kalmaz.
- [x] B2 — CMAF (fMP4) segment üretimi + HLS manifesti (birincil); DASH manifesti aynı segmentlerden opsiyonel | Kabul: tarayıcıda hls.js ile oynatılabilir çıktı. Bkz. [`../tech-stack.md`](../tech-stack.md) §2.
- [x] B3 — Adaptif bitrate (çoklu kalite kademesi) | Kabul: kademeler üretilir, geçiş çalışır.
- [x] B4 — Transcode oturum yönetimi + segment önbelleği/temizliği | Kabul: disk şişmez, oturum kapanınca temizlenir.

### Epik C: Ses & altyazı
- [x] C1 — Ses kanalı/dil seçimi + gerekirse downmix | Kabul: çoklu ses akışı seçilebilir.
- [x] C2 — Altyazı işleme (gömülü çıkarma, harici dosya, gerekirse burn-in) | Kabul: altyazı görüntülenir.

### Epik D: Streaming uçları
- [x] D1 — Streaming API uçları (manifest, segment, oturum) + yetki kontrolü | Kabul: yetkisiz akış engellenir.
- [x] D2 — Seek/arama ve devam (resume) desteği | Kabul: ileri/geri sarma çalışır.

## Kabul Kriterleri (Sprint Çıktısı)
- Bir medya dosyası tarayıcıda (direct play veya CPU transcode ile) baştan sona oynatılır.
- Ses/altyazı seçimi ve seek çalışır; transcode oturumları temiz yönetilir.

## Riskler
- **En yüksek teknik risk.** Kodek/konteyner çeşitliliği geniş → uyumluluk matrisi ve testler kritik.
- ffmpeg süreç/kaynak yönetimi hatası sistemi etkiler → sıkı sınırlar.

## Kapsam Dışı
- Donanım hızlandırma — Sprint 07 (bu sprint yalnızca CPU).
