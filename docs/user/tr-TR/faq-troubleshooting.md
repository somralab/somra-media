# SSS ve Sorun Giderme

## Kurulum

**Konteyner hemen kapanıyor** — günlükleri kontrol edin: `docker logs somra`. `SOMRA_DATA_DIR` yazılabilir olmalı.

**Arayüze erişilemiyor** — port eşlemesini ve güvenlik duvarını doğrulayın. Sağlık: `curl localhost:8080/api/v1/health`.

## Kütüphane

**Tarama dosya bulamıyor** — yolun konteyner içinde salt okunur bağlandığını ve uzantıların desteklendiğini doğrulayın.

**Metadata eksik** — `SOMRA_TMDB_API_KEY` ayarlayın veya geliştirmede test metadata kullanın.

## Oynatma

**ffmpeg bulunamadı** — üretim imajında ffmpeg vardır; yerel `go run` için ffmpeg kurulu olmalı.

**Transcode başarısız** — günlükleri inceleyin; önbellek biriminde boş alan olduğundan emin olun.

## Veritabanı

**SQLite kilitli** — aynı `data/` dizinine yalnızca bir Somra örneği yazmalı. Yinelenen konteynerleri durdurun.

## GPU

[gpu-setup.md](../gpu-setup.md) dosyasına bakın. `/api/v1/system/detect` hızlandırıcıları rapor etmeli.

## Eklentiler

**Bağlantı testi başarısız** — eklenti URL'si, kimlik bilgileri ve konteyner ağını kontrol edin.

## Yardım

- [GitHub Issues](https://github.com/somralab/somra-media/issues)
- Güvenlik açıkları için [SECURITY.md](../../SECURITY.md)
