# Oynatma

Somra medyayı tarayıcıda HLS (CMAF) ile oynatır. Sunucu, istemci codec'leri destekliyorsa **doğrudan oynatma**, aksi halde **transcode** seçer.

## Oynatmayı başlatma

1. **Kütüphaneler** veya **Ana sayfa**dan bir başlık açın.
2. **Oynat** düğmesine tıklayın.
3. Oynatıcı otomatik yüklenir; klavye kısayollarını kullanın (boşluk, oklar, F tam ekran).

## Transcode

- Yazılım transcode ffmpeg kullanır; eşzamanlı oturumları **Ayarlar → Oynatma** altında sınırlayın.
- Donanım hızlandırma (QSV/NVENC/VAAPI) etkinleştirildiğinde otomatik algılanır — [gpu-setup.md](../gpu-setup.md).

## Altyazılar

Detay sayfasında altyazı arayın ve ekleyin. Gömülü altyazılar transcode gerektirir.

## Sorun giderme

- **Dönen simge bitmiyor:** `/api/v1/health` → `ffmpeg` ve `transcode` kontrollerine bakın.
- **Tamponlama:** kaliteyi düşürün veya HW hızlandırmayı etkinleştirin; önbellek disk alanını doğrulayın.
