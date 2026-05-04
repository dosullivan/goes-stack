# goes-api-go

## Description
A lightweight and fast API for retrieving [GOES satellite data](https://www.goes-r.gov/) stored in an S3 bucket, written in Go. This service is designed to be used in conjunction with goes-web, a lightweight web frontend for viewing GOES satellite data. The service is meant to run in a local network alongside a locally-hosted S3-compatible bucket (e.g. RustFS or MinIO). The images are notably provided through a `/proxy/image` endpoint, which streams the image from the bucket through the api. With this setup, the frontend application goes-viewer, also running locally, can be served through a cloudflare tunnel or tailscale funnel on the internet, and the end user can pull the image through.

## Motivation
Basically just wanted to have an easier API to use than the S3 API with a frontend react app, while also making the proxying mentioned above possible, to minimize the surface area exposed to the internet.

## Requirements
This service assumes that you have an S3 bucket that is filled with GOES satellite data, likely from [goesrecv](https://pietern.github.io/goestools/commands/goesrecv.html), a system for capturing and processing the satellite data locally. It assumes that you have the same directory structure as the data folder used by `goesproc` as it's writing data. Basically, this implies using some cron script to simply copy the data from the data folder used by `goesproc` (e.g. goes16) to an S3 bucket.

You can find an [example script](./examples/upload.sh) for syncing the data from the data folder used by `goesproc` to an S3 bucket.

This project was developed against RustFS, MinIO, and DigitalOcean Spaces, but should work with any S3-compatible storage.

## Configuration
The following environment variables are required:
- `ACCESS_KEY_ID` - The access key for the S3 bucket.
- `SECRET_ACCESS_KEY` - The secret access key for the S3 bucket.
- `S3_ENDPOINT` - The endpoint for the S3 API.
- `BUCKET_NAME` - The name of the S3 bucket.
- `USE_SSL_FOR_S3` - Whether to use SSL for S3. Set to `false` for plain HTTP local setups (e.g. RustFS/MinIO without TLS).
- `TRUSTED_PROXIES`- A comma-separated list of trusted proxies, in case you want to set them. This isn't used for anything yet.

## API Usage
The following API endpoints are available:

- `/latest` — Returns the latest image URL.
```shell
curl http://localhost:3000/latest
# => { "imageUrl": "http://localhost:9000/.../some-latest-image.png" }
```

- `/available-dates` — Returns a list of dates with available images.
```shell
curl http://localhost:3000/available-dates
# => { "availableDates": ["2024-08-23", "2024-08-22", "2024-08-21"] }
```

- `/archive/{date}` — Returns the image URLs for a given date (YYYY-MM-DD, CST-based selection).
```shell
curl http://localhost:3000/archive/2024-05-16
# => {
#   "imageUrls": [
#     "http://localhost:9000/.../2024-05-16/....png",
#     "http://localhost:9000/.../2024-05-16/....png"
#   ]
# }
```

- `/weather/products` — Returns the list of available weather product keys and metadata.
```shell
curl http://localhost:3000/weather/products
# => { "products": [ { "key": "fd_color", "title": "GOES-19 Full Disk Color", "category": "goes19_fd" }, ... ] }
```

- `/weather/products/{product}` — Returns images for a product. Optional `date=YYYY-MM-DD` (UTC) query.
```shell
curl "http://localhost:3000/weather/products/fd_color"
# or
curl "http://localhost:3000/weather/products/fd_color?date=2024-08-23"
# => { "product": "fd_color", "images": [ { "url": "http://localhost:9000/...png", "timestamp": "2024-08-23T12:34:56Z", "filename": "...png" } ], "count": 42 }
```

- `/emwin/text/categories` — Returns EMWIN text categories.
```shell
curl http://localhost:3000/emwin/text/categories
# => { "categories": [ { "key": "weather_warnings", "title": "Weather Warnings & Watches" }, ... ] }
```

- `/emwin/text/files` — Returns EMWIN text files. Optional queries: `category`, `date=YYYY-MM-DD`, `station`. Limited to 100 items.
```shell
curl "http://localhost:3000/emwin/text/files?category=forecasts&date=2024-08-23"
# => { "files": [ { "url": "http://localhost:9000/emwin/...TXT", "timestamp": "2024-08-23T01:23:45-05:00", "filename": "...TXT", "productCode": "ASUS41", "station": "KGYX" } ], "count": 12, "filters": { "category": "forecasts", "station": "", "date": "2024-08-23" } }
```

- `/emwin/text/content` — Returns content for a specific EMWIN text file. Requires `key` query with the object key.
```shell
curl "http://localhost:3000/emwin/text/content?key=emwin/2024-08-23/ABC_DEF_20240823120000_...TXT"
# => { "objectKey": "emwin/2024-08-23/...TXT", "content": "...file contents..." }
```

- `/proxy/image` — Proxies an image from the configured S3 base URL through this API. Requires `url` query.
```shell
curl "http://localhost:3000/proxy/image?url=http://localhost:9000/bucket/path/to/image.png" --output image.png
# Streams the image through the API and saves it locally
```

## Deployment
(to be written)

## Development
This repo uses [direnv](https://direnv.net/) to manage environment variables for development. The .envrc file is a sample with default values. It's setup to pick up overriding values from `.envrc.local` if it exists. You can copy the `.envrc` file to `.envrc.local` and modify it to your needs. The `.envrc.local` file is ignored by git.

Likewise, a sample `docker-compose.yml` file is provided for running the server in a container locally. You can copy this file to `docker-compose.override.yml` and modify it to your needs. The `docker-compose.override.yml` file is ignored by git.

To start the server in development mode, just run `make run`. This will start the server on port 3000.

To test your docker compose setup, run `make local-build-deploy`.

## improvements for the future:
- EMWIN and NWS text retrieval?