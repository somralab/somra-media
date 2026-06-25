# Kurulum

Somra tek bir Docker imajı olarak gelir. Ayrı veritabanı veya ters vekil sunucu gerekmez.

## Hızlı başlangıç (tek satır)

v1.0.0 yayımlandıktan sonra:

```bash
docker run -d --name somra -p 8080:8080 \
  -e SOMRA_JWT_SECRET="$(openssl rand -base64 32)" \
  -v somra-data:/data -v somra-cache:/cache \
  -v /media/yolunuz:/media:ro \
  ghcr.io/somralab/somra-media:1.0.0
```

**http://localhost:8080** adresini açın ve kurulum sihirbazını tamamlayın.

## Docker Compose (önerilen)

Üretim örneği: [`deploy/docker-compose.production.yml`](../../deploy/docker-compose.production.yml)

```bash
mkdir -p deploy/config deploy/data deploy/cache deploy/media
cp deploy/.env.production.example deploy/.env
# deploy/.env dosyasını düzenleyin — SOMRA_JWT_SECRET ayarlayın
docker compose -f deploy/docker-compose.production.yml up -d
```

Kaynaktan geliştirme derlemesi:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

## Ortam değişkenleri

| Değişken | Zorunlu | Açıklama |
|----------|---------|----------|
| `SOMRA_JWT_SECRET` | Üretimde evet | JWT imzalama için 32+ karakter gizli anahtar |
| `SOMRA_DATA_DIR` | Hayır | SQLite ve durum (varsayılan `/data`) |
| `SOMRA_CACHE_DIR` | Hayır | Transcode önbelleği (varsayılan `/cache`) |
| `SOMRA_HTTP_ADDR` | Hayır | Dinleme adresi (varsayılan `:8080`) |

## GPU geçişi

VAAPI, NVIDIA ve QSV katmanları için [`docs/gpu-setup.md`](../gpu-setup.md) dosyasına bakın.

## HTTPS (ters vekil)

Somra'yı yerel ağınızda veya alan adınızda Caddy veya Traefik arkasına alın:

```caddy
somra.ornek.com {
  reverse_proxy localhost:8080
}
```

Somra'yı TLS olmadan doğrudan internete açmayın.

## Sistem gereksinimleri

- **CPU:** 4 çekirdek önerilir
- **RAM:** 8 GB önerilir
- **Disk:** `/data` ve `/cache` için SSD; medya ayrı depolamada
- **OS:** Docker 24+ destekleyen Linux (amd64 veya arm64)
