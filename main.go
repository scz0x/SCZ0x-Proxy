package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

const (
	concurrencyLimit = 200
	timeoutSeconds   = 4
	testURL          = "https://www.google.com"
)

var (
	httpFile   *os.File
	socks4File *os.File
	socks5File *os.File
	summaryLog *os.File
	errorLog   *os.File

	successCount  = 0
	totalChecked  = 0
	totalToCheck  = 0
	mu            sync.Mutex
	startTime     time.Time
	outputDir     string
)

func initLogger() {
	os.MkdirAll("logs", 0755)
	logF, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Can't open error log:", err)
		os.Exit(1)
	}
	log.SetOutput(logF)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	errorLog = logF
}

func initOutputFiles() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	outputDir = filepath.Join("results", timestamp)
	os.MkdirAll(outputDir, 0755)

	httpFile, _ = os.OpenFile(filepath.Join(outputDir, "http.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	socks4File, _ = os.OpenFile(filepath.Join(outputDir, "socks4.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	socks5File, _ = os.OpenFile(filepath.Join(outputDir, "socks5.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	summaryLog, _ = os.OpenFile(filepath.Join(outputDir, "summary.log"), os.O_CREATE|os.O_WRONLY, 0644)
}

func writeProxy(proxyStr, proxyType string) {
	var f *os.File
	switch proxyType {
	case "http", "https":
		f = httpFile
	case "socks4":
		f = socks4File
	case "socks5":
		f = socks5File
	default:
		return
	}
	f.WriteString(proxyStr + "\n")
}

func checkGoogle(proxyStr, proxyType string, sem chan struct{}) {
	defer func() { <-sem }()
	client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}

	switch proxyType {
	case "http", "https":
		proxyURL, _ := url.Parse(fmt.Sprintf("%s://%s", proxyType, proxyStr))
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	case "socks4", "socks5":
		dialer, err := proxy.SOCKS5("tcp", proxyStr, nil, proxy.Direct)
		if err != nil {
			increment()
			return
		}
		client.Transport = &http.Transport{Dial: dialer.Dial}
	default:
		increment()
		return
	}

	req, _ := http.NewRequest("GET", testURL, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		increment()
		return
	}
	resp.Body.Close()

	mu.Lock()
	successCount++
	writeProxy(proxyStr, proxyType)
	mu.Unlock()

	increment()
}

func increment() {
	mu.Lock()
	totalChecked++
	printProgress()
	mu.Unlock()
}

func printProgress() {
	elapsed := time.Since(startTime).Seconds()
	if elapsed == 0 {
		elapsed = 1
	}
	speed := float64(totalChecked) / elapsed
	fmt.Printf("\rProgress: %d / %d | ✅ Working: %d | ⚡ %.1f req/s", totalChecked, totalToCheck, successCount, speed)
}

func detectType(proxyStr string) string {
	port := strings.Split(proxyStr, ":")
	if len(port) != 2 {
		return "unknown"
	}
	switch port[1] {
	case "1080":
		return "socks5"
	case "1081":
		return "socks4"
	case "8080", "3128", "8000", "8888", "80":
		return "http"
	default:
		return "http"
	}
}

func fetchProxiesFrom(api string) ([]string, error) {
	resp, err := http.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proxies []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && strings.Contains(line, ":") {
			proxies = append(proxies, line)
		}
	}
	return proxies, nil
}

func writeSummary(elapsed float64) {
	summary := fmt.Sprintf("Total Proxies: %d\nWorking: %d\nTime: %.2f seconds\nSpeed: %.2f req/s\n",
		totalToCheck, successCount, elapsed, float64(totalChecked)/elapsed)
	summaryLog.WriteString(summary)
}

func main() {
	initLogger()
	initOutputFiles()
	defer errorLog.Close()
	defer httpFile.Close()
	defer socks4File.Close()
	defer socks5File.Close()
	defer summaryLog.Close()
	os.MkdirAll("results", 0755)

	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║       SCZ-Proxy by Ahmed ⚡       ║")
	fmt.Println("║     Ultra-Fast Proxy Checker     ║")
	fmt.Println("╚══════════════════════════════════╝")

	// Count total APIs before reading
	apiLines, err := os.ReadFile("apis.txt")
	if err != nil {
		log.Fatalf("Cannot open apis.txt: %v", err)
	}
	lines := strings.Split(string(apiLines), "\n")
	totalAPIs := 0
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			totalAPIs++
		}
	}
	fmt.Printf("🔄 Fetching proxies from %d source(s)...\n", totalAPIs)

	// Re-open for processing
	apisFile, _ := os.Open("apis.txt")
	defer apisFile.Close()

	var allProxies []string
	scanner := bufio.NewScanner(apisFile)
	for scanner.Scan() {
		api := strings.TrimSpace(scanner.Text())
		if api == "" || strings.HasPrefix(api, "#") {
			continue
		}
		proxies, err := fetchProxiesFrom(api)
		if err == nil {
			allProxies = append(allProxies, proxies...)
		}
	}

	unique := make(map[string]bool)
	for _, p := range allProxies {
		unique[p] = true
	}
	var proxyList []string
	for p := range unique {
		proxyList = append(proxyList, p)
	}

	totalToCheck = len(proxyList)
	fmt.Printf("🔎 Total unique proxies to check: %d\n", totalToCheck)

	semaphore := make(chan struct{}, concurrencyLimit)
	startTime = time.Now()

	for _, proxy := range proxyList {
		semaphore <- struct{}{}
		go checkGoogle(proxy, detectType(proxy), semaphore)
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	elapsed := time.Since(startTime).Seconds()
	writeSummary(elapsed)
	fmt.Printf("\n\n✅ Done! Working: %d / %d\n", successCount, totalToCheck)
	fmt.Printf("🗂 Results in: %s\n", outputDir)
	fmt.Printf("⏱️ Total Time: %.2f seconds\n", elapsed)
	fmt.Print("\nPress ENTER to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}