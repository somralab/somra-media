# Kütüphane

## Tarama

Somra kütüphane köklerini izler ve planlı tam/artımlı taramalar çalıştırır. Dosya adları başlık, yıl, sezon ve bölüm için ayrıştırılır.

- **Tam tarama:** tüm yolları yeniden tarar (toplu içe aktarmadan sonra).
- **Artımlı tarama:** yalnızca yeni/değişen dosyaları algılar.

Tarama ilerlemesi arayüze aktarılır. Büyük kütüphanelerde ilk tarama birkaç dakika sürebilir.

## Metadata

Eşleşen başlıklar poster, açıklama ve oyuncu kadrosunu yapılandırılmış sağlayıcılardan gösterir (TMDB API anahtarı ayarlandığında). Eşleşmeyen öğeler detay sayfasından yeniden eşleştirilebilir.

## Dosya düzeni ipuçları

```
movies/Film Adı (2020)/Film Adı (2020).mkv
tv/Dizi Adı/Season 01/Dizi Adı S01E01.mkv
```

Ayrıştırıcı örnekleri için `testdata/library/README.md` dosyasına bakın.

## İzinler

- **library:read** — göz atma ve oynatma
- **library:write** — kütüphane oluşturma, tarama, yeniden eşleştirme

Yöneticiler varsayılan olarak tüm izinlere sahiptir.
