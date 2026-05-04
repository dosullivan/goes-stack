# goesproc-docker

Containerized [goesproc](https://pietern.github.io/goestools/commands/goesproc.html) setup that turns the demodulated TCP stream from satdump (running on the Pi in the yard) into images and weather products on disk, then ships them to RustFS (or any S3-compatible backend).

## Services

The stack defined in [`docker-compose.yml`](docker-compose.yml) runs three containers, all coordinated through the same `./data/` volume.

### `goesproc`
Built from [`Dockerfile`](Dockerfile) — clones and compiles [goestools](https://github.com/pietern/goestools) on top of `ubuntu:22.04`. At runtime it subscribes to the satdump TCP stream on the Pi (`tcp://<pi-ip>:5004` — set in your override file) and writes decoded products into `./data/` according to the rules in [`config/goesproc.conf`](config/goesproc.conf).

### `goesproc-watchdog`
A tiny Alpine sidecar that watches the Pi's TCP port and restarts the `goesproc` container whenever the stream comes back after being down. Without this, a brief network blip on the Pi can leave goesproc stuck in a bad state until manually restarted.

The watchdog logic lives in [`scripts/watch.sh`](scripts/watch.sh) and is mounted read-only into the container. Tunable knobs (`CHECK_INTERVAL`, `UP_STREAK`, `DOWN_STREAK`, `COOLDOWN_SEC`) are set via env vars in `docker-compose.yml`.

### `uploader`
Built from [`Dockerfile.uploader`](Dockerfile.uploader) — an Alpine container with the AWS CLI. Every `UPLOAD_INTERVAL` seconds (15 minutes by default) it runs [`scripts/upload.sh`](scripts/upload.sh), which:

1. Calls [`scripts/preprocess_emwin.sh`](scripts/preprocess_emwin.sh) to fix EMWIN files that goesproc landed in the `1969-12-31` folder (a known goesproc quirk where files arrive without a usable timestamp header — we re-derive the date from the filename and move them into the correct `YYYY-MM-DD` directory before upload).
2. Runs `aws s3 sync` to push everything in `./data/` into the configured bucket.
3. Removes local files that successfully made it remote (verified with `aws s3api head-object`), so the local disk doesn't fill up indefinitely.

By default `aws s3 sync` runs without `--delete`, so files deleted remotely won't be deleted locally on the next sync. Set `ALLOW_REMOTE_DELETIONS=true` in your override if you want bidirectional sync.

Logs from each upload cycle are appended to `./logs/upload.log`.

## Configuration

Everything user-specific (Pi IP, S3 endpoint, credentials, bucket name, upload interval) lives in [`docker-compose.override.yml`](docker-compose.override.yml), which is gitignored. The base `docker-compose.yml` ships with `pi.local` placeholders for the Pi's address — your override should replace them with the real IP.

Example override (see the top-level [README](../README.md#2-start-goesproc-processing-with-auto-upload) for the full template):

```yaml
services:
  goesproc:
    command: ["goesproc", "-c", "/config/goesproc.conf", "-m", "packet", "--subscribe", "tcp://192.168.1.42:5004"]

  goesproc-watchdog:
    environment:
      TARGET_HOST: "192.168.1.42"

  uploader:
    environment:
      - S3_ENDPOINT=rustfs.local:9000
      - S3_ACCESS_KEY=...
      - S3_SECRET_KEY=...
      - BUCKET_NAME=goes-data
      - UPLOAD_INTERVAL=900
```

## Running

```bash
docker-compose up -d                       # start the whole stack
docker-compose logs -f goesproc            # tail goesproc decoder output
docker-compose logs -f goesproc-watchdog   # tail watchdog state changes
tail -f logs/upload.log                    # tail upload cycles
```

## Customizing what gets decoded

Edit [`config/goesproc.conf`](config/goesproc.conf) to add, remove, or reconfigure handlers. Each `[[handler]]` block defines a product type, output directory, and (for images) any color gradients/maps applied. Restart the `goesproc` container for changes to take effect.
