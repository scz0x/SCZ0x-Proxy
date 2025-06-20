# SCZ-Proxy v1

A blazing-fast proxy scanner written in Go. Supports HTTP, SOCKS4, and SOCKS5 protocols with smart concurrency, flexible input sources, and real-time performance tracking.

## ⚙ Features

- Supports `HTTP`, `SOCKS4`, and `SOCKS5` proxies
- Input from APIs (`apis.txt`), text files (`proxies.txt`), or folders (`sources/`)
- Filter proxies by type using `--only=http|socks5|socks4`
- Set request timeout with `--timeout`
- Silent mode for automation: `--silent`
- Real-time progress and request speed tracking
- Saves working proxies by type only when results are found

## 🚀 Usage

```bash
# Build
go build -o scz-proxy

# Scan from APIs
./scz-proxy -mode=api

# Scan from file and filter SOCKS5
./scz-proxy -mode=txt -only=socks5

# Scan folder with custom timeout
./scz-proxy -mode=folder -timeout=6
```

Ensure these files/folders are in place:

- `apis.txt`: one proxy API link per line
- `proxies.txt`: plain list of proxies (IP:PORT)
- `sources/`: folder containing multiple .txt files

## 📦 Output

Working proxies are saved to:

```
results/YYYY-MM-DD_HH-MM-SS/
├── http.txt
├── socks4.txt
├── socks5.txt
└── summary.log
```

## 📢 Stay Updated

Join the official Telegram channel for updates, APIs, discussions, and releases:  
👉 [https://t.me/SCZ0X_CH]

---
