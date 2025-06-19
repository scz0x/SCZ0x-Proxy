
# 🚀 SCZ-Proxy

**SCZ-Proxy** is a blazing-fast proxy validator written in Go. It supports HTTP, SOCKS4, and SOCKS5 protocols, with concurrent validation, Google reachability testing, deduplication, and auto-organized results — all with a clean terminal interface.

Built for developers and toolmakers who value speed and clarity.

---

## 🎯 Features

- ✅ High-speed concurrent proxy checking
- 📡 Google reachability test per proxy
- 🧠 Auto deduplication before scan
- 🧪 Supports HTTP, SOCKS4, SOCKS5
- ⏱️ Live progress bar with real-time stats
- 📁 Output saved in timestamped folders
- 📊 Summary report generated after each run

---

## ⚙️ Requirements

- [Go 1.20+](https://go.dev/dl/)
- Terminal or CMD (cross-platform)
- `apis.txt` → list of proxy source URLs (one per line)

No external Go packages required.

---

## 🚀 Getting Started

1. Clone the repo:
   ```bash
   git clone https://github.com/scz0x/SCZ0x-Proxy.git
   cd SCZ0x-Proxy
   ```

2. Prepare Go modules:
   ```bash
   go mod tidy
   ```

3. Add proxy sources in `apis.txt`:
   ```txt
   https://api.proxyscrape.com/v2/?request=getproxies&protocol=http
   https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt
   ```

4. Build:
   ```bash
   go build -o scz-proxy
   ```

5. Run:
   ```bash
   ./scz-proxy     # or scz-proxy.exe on Windows
   ```

---

## 📁 Output Example

Each run creates a new folder in `/results/` like:

```
results/2025-06-19_22-50-12/
├── http.txt
├── socks4.txt
├── socks5.txt
└── summary.log
```

And all errors go to `/logs/error.log`.

---

## 📋 Output Sample

```
🔄 Fetching proxies from 15 source(s)...
🔎 Total unique proxies to check: 1042

Progress: 482 / 1042 | ✅ Working: 121 | ⚡ 73.2 req/s
```

---

## 🧱 Project Layout

```
SCZ-Proxy/
├── main.go
├── apis.txt
├── README.md
├── go.mod / go.sum
├── .gitignore
├── /results/
└── /logs/
```

Crafted by [Scz0x](https://github.com/scz0x) — SCZ-Proxy is designed to help you test more and wait less.
---
