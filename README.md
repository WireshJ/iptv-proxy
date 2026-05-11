# 📡 IPTV Proxy

A self-hosted reverse proxy for IPTV M3U playlists and Xtream Codes API. Expose your IPTV provider behind your own credentials and hostname, with optional VPN routing via Gluetun.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-GPL--3.0-green)
![Docker](https://img.shields.io/badge/docker-ghcr.io%2Fwireshj%2Fiptv--proxy-blue?logo=docker)

---

## Table of Contents

- [Features](#-features)
- [Requirements](#-requirements)
- [Installation](#-installation)
  - [Docker (recommended)](#docker-recommended)
  - [Docker Compose with Gluetun VPN](#docker-compose-with-gluetun-vpn)
  - [Manual (Go binary)](#manual-go-binary)
- [Configuration](#-configuration)
  - [Environment variables](#environment-variables)
- [Usage](#-usage)
  - [M3U proxy](#m3u-proxy)
  - [Xtream Codes proxy](#xtream-codes-proxy)
- [Credits](#-credits)

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🔄 **M3U proxy** | Rewrites all channel URLs in your M3U playlist to point to the proxy |
| 📺 **M3U8 proxy** | Transparent proxy for HLS (m3u8) streams |
| 🎬 **Xtream Codes** | Full proxy for the Xtream Codes client API — live, VOD, series, EPG |
| 🔐 **Credential swap** | Replaces provider credentials with your own user/password |
| 🌐 **VPN support** | Route all traffic through Gluetun (or any network namespace) |
| ⚡ **Caching** | M3U playlist cached locally to reduce provider requests |

---

## 📋 Requirements

| Requirement | Notes |
|-------------|-------|
| Docker | For the recommended container setup |
| Go 1.22+ | Only for manual/binary builds |
| IPTV provider | M3U URL or Xtream Codes credentials |

---

## 🚀 Installation

### Docker (recommended)

```bash
docker run -d \
  --name iptv-proxy \
  -p 8080:8080 \
  -e HOSTNAME=your-server-ip \
  -e PORT=8080 \
  -e M3U_URL="http://provider.example.com:8000/get.php?username=user&password=pass&type=m3u_plus&output=mpegts" \
  -e USER=myuser \
  -e PASSWORD=mypassword \
  -e GIN_MODE=release \
  ghcr.io/wireshj/iptv-proxy:latest
```

**Port:** `8080`

---

### Docker Compose with Gluetun VPN

Download the compose file and fill in your credentials:

```bash
mkdir -p /data/apps/iptv-proxy
curl -o /data/apps/iptv-proxy/docker-compose.yml \
  https://raw.githubusercontent.com/WireshJ/iptv-proxy/master/docker-compose.yml

cd /data/apps/iptv-proxy
# Edit docker-compose.yml with your VPN and IPTV credentials
docker compose up -d
```

> The `iptv-proxy` container shares the Gluetun network namespace — all IPTV traffic is routed through the VPN automatically.

**Volumes:**
- `/data/apps/gluetun` — Gluetun configuration and state
- `/data/apps/iptv-proxy` — optional local M3U file mount

---

### Manual (Go binary)

```bash
git clone https://github.com/WireshJ/iptv-proxy.git
cd iptv-proxy

go build -mod=vendor -o iptv-proxy .

./iptv-proxy \
  --hostname your-server-ip \
  --port 8080 \
  --m3u-url "http://provider.example.com/iptv.m3u" \
  --user myuser \
  --password mypassword
```

All CLI flags can also be set as environment variables (replace `-` with `_`, e.g. `M3U_URL`).

---

## ⚙️ Configuration

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOSTNAME` | *(required)* | Hostname or IP that clients will connect to |
| `PORT` | `8080` | Port the proxy listens on |
| `ADVERTISED_PORT` | same as `PORT` | Port exposed to clients — useful behind a reverse proxy |
| `M3U_URL` | — | Remote M3U URL or path to local file (e.g. `/root/iptv/iptv.m3u`) |
| `M3U_FILE_NAME` | `iptv.m3u` | Filename of the proxified M3U endpoint |
| `M3U_CACHE_EXPIRATION` | `1` | M3U cache lifetime in hours |
| `USER` | `usertest` | Username clients use to authenticate with the proxy |
| `PASSWORD` | `passwordtest` | Password clients use to authenticate with the proxy |
| `XTREAM_USER` | — | Xtream Codes username from your provider |
| `XTREAM_PASSWORD` | — | Xtream Codes password from your provider |
| `XTREAM_BASE_URL` | — | Xtream Codes base URL e.g. `http://provider.example.com:8080` |
| `XTREAM_API_GET` | `false` | Generate `get.php` from the Xtream API instead of the original endpoint |
| `HTTPS` | `false` | Use HTTPS for proxified URLs |
| `CUSTOM_ENDPOINT` | — | Optional path prefix for all proxy endpoints |
| `GIN_MODE` | — | Set to `release` to suppress Gin debug output |
| `DEBUG` | — | Set to `true` to enable debug logging |

> **Xtream auto-detection:** If your `M3U_URL` contains `/get.php` with `username` and `password`, Xtream mode is enabled automatically — no need to set `XTREAM_USER`, `XTREAM_PASSWORD`, or `XTREAM_BASE_URL` separately.

---

## 🔌 Usage

### M3U proxy

After starting the proxy, your proxified M3U playlist is available at:

```
http://<HOSTNAME>:<PORT>/iptv.m3u?username=<USER>&password=<PASSWORD>
```

All channel URLs in the playlist point back to the proxy. Clients stream through the proxy to the original provider.

**Example with original M3U:**
```
#EXTM3U
#EXTINF:-1 tvg-name="Channel 1",Channel 1
http://provider.example.com:1234/user/pass/1
```

**Proxified output:**
```
#EXTM3U
#EXTINF:-1 tvg-name="Channel 1",Channel 1
http://your-server-ip:8080/<token>/myuser/mypassword/0/1
```

---

### Xtream Codes proxy

When Xtream mode is active, the proxy exposes the full Xtream Codes API on your hostname with your credentials:

| Endpoint | Description |
|----------|-------------|
| `GET /get.php` | M3U playlist |
| `GET /player_api.php` | Live, VOD, series, EPG data |
| `POST /player_api.php` | Same, for apps that POST |
| `GET /xmltv.php` | EPG XML data |
| `GET /live/<user>/<pass>/:id` | Live stream |
| `GET /movie/<user>/<pass>/:id` | VOD stream |
| `GET /series/<user>/<pass>/:id` | Series stream |

**Original credentials (provider):**
```
user:     xtream_user
password: xtream_password
base-url: http://provider.example.com:8080
```

**Proxified credentials (clients use these):**
```
user:     myuser
password: mypassword
base-url: http://your-server-ip:8080
```

---

## 🙏 Credits

Originally created by [Pierre-Emmanuel Jacquier](https://github.com/pierre-emmanuelJ) — [pierre-emmanuelJ/iptv-proxy](https://github.com/pierre-emmanuelJ/iptv-proxy).

Extended by [incmve](https://github.com/incmve) — [incmve/iptv-proxy](https://github.com/incmve/iptv-proxy):
- Fixed Xtream Codes EPG not loading
- Fixed Xtream Codes VOD (Shows & Movies) with incomplete provider responses
- Continue-on-error for malformed `EXTINF` entries
- Gluetun VPN integration

Further developed by [WireshJ](https://github.com/WireshJ):
- Updated to Go 1.22, fixed deprecated APIs, shared HTTP client
- Multi-platform Docker image (`linux/amd64`, `linux/arm64`) via GHCR
- Fixed GitHub Actions workflow, updated all references

**Powered by:**
- [gin](https://github.com/gin-gonic/gin)
- [cobra](https://github.com/spf13/cobra)
- [go.xtream-codes](https://github.com/tellytv/go.xtream-codes)
- [gluetun](https://github.com/qdm12/gluetun)
