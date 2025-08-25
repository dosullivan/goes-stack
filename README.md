# GOES-stack

A modern, containerized solution for processing, storing, and viewing GOES satellite imagery + data. This monorepo contains all components needed to run a complete GOES satellite ground station.

## Architecture Overview

```
Raspberry Pi (Yard) → Processing Server → MinIO Storage → Web Frontend
    ↓                      ↓                 ↓              ↓
  satdump/goesrecv    →  goesproc-docker  →  goes-api-go  →  goes-viewer
```

### Reasoning
While you can run tools like [satdump](https://www.satdump.org/) and have this all in one, I wanted to take a different approach. I wanted the Raspberry Pi in the yard to have the least power draw possible as it is solar powered. Additionally, I wanted each part of the system to be isolated. 

### Recommended hardware
- A Raspberry Pi (ideally 4b, as it has enough processing power + less power draw than a 5) to put in a weatherproof enclosure
- [Nooelec NESDR SMArTee XDR SDR](https://www.nooelec.com/store/sdr/sdr-receivers/smart/nesdr-smartee-xtr-sdr.html)
- [Discovery Dish + L-Band Satellite Feed](https://www.crowdsupply.com/krakenrf/discovery-dish)
- A modern desktop/server inside (connected to the same network as the pi, via wifi or ethernet)
- Internet connection (you can expose the frontend with something like Cloudflare tunnel).

### Recommended software for running on the pi in the yard
[goesrecv](https://github.com/pietern/goestools) and [satdump](https://github.com/SatDump/SatDump) work equally well, but I'm using satdump.


## Components

### 🛰️ goes-viewer
Modern Next.js 15 frontend for viewing GOES satellite imagery with a clean, responsive interface.

**Features:**
- Latest image viewer with automatic refresh
- Image archive with date picker
- Modern UI built with Shadcn/UI and Radix primitives
- TypeScript for type safety
- Dark/light theme support

### 🔧 goes-api-go
Lightweight Go API server that provides RESTful endpoints for accessing satellite images stored in MinIO/S3-compatible storage.

**Endpoints:**
- `/latest` - Get the most recent satellite image
- `/archive/:date` - Get images for a specific date
- `/available-dates` - List all dates with available imagery
- `/proxy/image` - Secure image proxy with URL validation

### 🐳 goesproc-docker
Containerized goesproc setup for processing raw GOES satellite data from TCP streams.

**Features:**
- Processes raw GOES data from satdump/goesrecv TCP streams
- Automatic image upload to MinIO every 15 minutes via sidecar container
- Local cleanup of successfully uploaded files
- Comprehensive logging and error handling
- Environment-based configuration

## Quick Start

### 1. Set up MinIO Storage

```bash
cd minio
docker-compose up -d
```

This creates an S3-compatible storage backend at `http://localhost:9000`.

### 2. Configure Raspberry Pi (Yard Setup)

On your Raspberry Pi with RTL-SDR, run satdump to receive GOES-18:

```bash
satdump live goes_hrit_tcp GOES_18 \
  --source rtlsdr --samplerate 2.4e6 --frequency 1694.1e6 \
  --agc --bias --fill_missing \
  --http_server 0.0.0.0:8080 \
  --tcp_port 5004
```

### 3. Start goesproc Processing with Auto-Upload

```bash
cd goesproc-docker
# Copy and configure environment
cp .env.example .env
# Edit .env with your MinIO endpoint and credentials
# Update docker-compose.yml to point to your Pi's IP for TCP stream
docker-compose up -d
```

The uploader sidecar will automatically:
- Upload processed images to MinIO every 15 minutes
- Clean up local files after successful upload
- Log all operations to `./logs/upload.log`

### 4. Start the API Server

```bash
cd goes-api-go
# Copy and configure environment
cp .env.example .env
# Edit .env with your MinIO credentials
docker-compose up -d
```

### 5. Start the Web Frontend

```bash
cd goes-viewer
npm install
npm run dev
```

Visit `http://localhost:3000` to view your satellite imagery!

## Configuration

### Environment Variables

**goes-api-go:**
```bash
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=goes-images
PORT=3010
```

**goesproc-docker:**
```bash
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
BUCKET_NAME=goes-19
UPLOAD_INTERVAL=900  # 15 minutes
```

### MinIO Setup

1. Access MinIO console at `http://localhost:9001`
2. Create a bucket named `goes-19` etc
3. Configure appropriate access policies

### goesproc Configuration

- Edit `config/goesproc.conf` to configure your processing pipeline
- Update `docker-compose.yml` to point to your satdump/goesrecv TCP stream
- Monitor upload logs: `docker-compose logs -f uploader`

## Development

### goes-viewer (Next.js)
```bash
cd goes-viewer
npm run dev          # Development server with hot reload
npm run build        # Production build
npm run lint         # ESLint checking
```

### goes-api-go
```bash
cd goes-api-go
go run main.go       # Development server
make build          # Build binary
make test           # Run tests
```

## Deployment

Each component can be deployed independently using Docker:

```bash
# API Server
cd goes-api-go && docker-compose up -d

# Frontend (build and serve)
cd goes-viewer && docker build -t goes-viewer . && docker run -p 3000:3000 goes-viewer

# Processing
cd goesproc-docker && docker-compose up -d
```

## Troubleshooting

### Upload Issues
- Check MinIO connectivity: `docker-compose exec uploader mc ls minio/`
- View upload logs: `docker-compose logs -f uploader`
- Verify bucket permissions in MinIO console

### Processing Issues
- Check TCP connection to Raspberry Pi
- Verify goesproc configuration file
- Monitor processing logs: `docker-compose logs -f goesproc`

## Contributing

This is a personal project, but contributions are welcome! Please open issues or pull requests.

## License

MIT License.

---

*Hosted from my yard, when there is enough sunlight* ☀️