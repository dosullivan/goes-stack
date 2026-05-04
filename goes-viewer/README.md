# goes-viewer-v2

This is a rewrite of the goes-viewer frontend I had previously written. This application is built with Next.js and Tailwind CSS. This is made to work with the [goes-api-go](https://github.com/dosullivan/goes-api-go) backend API, which provides a REST API to retrieve images from a bucket by date.

This is made with local hosting in mind -- the goes-api-go backend is intended to be ran locally alongside this frontend. I have a local object storage setup, which goes-api-go is configured to pull from.

This frontend and the goes-api-go backend are both hosted on my personal server. A separate Raspberry Pi in my yard runs [goesrecv](https://pietern.github.io/goestools/commands/goesrecv.html), which is a tool that captures images and other data from the NOAA satellites using a software defined radio. 

The goesrecv tool is configured to send data to a server running [goesproc](https://pietern.github.io/goestools/commands/goesproc.html), which actually processes the data and writes it to a local filesystem. In turn, a cron script on that server checks for newly created files and uploads them to a local object storage server. That object storage server is then accessed by the goes-api-go backend to retrieve the images, which are then served to the frontend (this application).

## Getting Started

To get started, run the development server:

```bash
npm run dev
```

## Required Environment Variables

- `NEXT_PUBLIC_API_URL`: The URL of the goes-api-go backend API.

## Image hostname allowlist

Next.js's `<Image>` component refuses to load images from hostnames not declared in [`next.config.ts`](next.config.ts) under `images.remotePatterns`. The hostname there must exactly match the host that goes-api-go is returning in its `imageUrl` / `imageUrls` responses (which is built from goes-api-go's `S3_ENDPOINT` env var).

The hostname currently checked into [`next.config.ts`](next.config.ts) (`s3.local`) is just a placeholder — replace it with whatever hostname your own S3 backend is reachable at before building.

If you change the S3 endpoint hostname (e.g. swap from `minio.local` to `rustfs.local`, move to a new domain, or stand up a second backend), you must:

1. Update the `hostname` entry in `next.config.ts`.
2. Rebuild the goes-viewer container (`docker compose up -d --build`) — `next.config.ts` is consumed at build time, not runtime.

Symptom of a mismatch: the API responses look right (`/latest` returns a sensible URL), but image tiles in the UI render blank or as broken-image placeholders. Browser devtools will show 400 responses from `/_next/image` with a message about the hostname not being allowed.
