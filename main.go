package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
)

const testURL = "https://www.google.com"

var (
	timeout       int
	onlyProtocol  string
	silent        bool
	outputDir     string
	successCount  int
	totalChecked  int
	totalToCheck  int
	startTime     time.Time
	mu            sync.Mutex
	httpFile      *os.File
	socks4File    *os.File
	socks5File    *os.File
	summaryLog    *os.File
	filesReady    bool
)

func main() {
	mode := flag.String("mode", "", "api / txt / folder")
	timeoutPtr := flag.Int("timeout", 4, "Timeout in seconds")
	only := flag.String("only", "", "Filter by proxy type")
	silentPtr := flag.Bool("silent", false, "No output")
	flag.Parse()

	timeout = *timeoutPtr
	onlyProtocol = strings.ToLower(*only)
	silent = *silentPtr

	if *mode == "" {
		fmt.Println("Choose source:")
		fmt.Println("1. API (apis.txt)")
		fmt.Println("2. TXT (proxies.txt)")
		fmt.Println("3. Folder (sources/)")
		fmt.Print("Your choice: ")
		var choice string
		fmt.Scanln(&choice)
		switch choice {
		case "1":
			*mode = "api"
		case "2":
			*mode = "txt"
		case "3":
			*mode = "folder"
		default:
			fmt.Println("Invalid choice.")
			return
		}
	}

	var allProxies []string
	switch *mode {
	case "api":
		allProxies = loadFromAPIs("apis.txt")
	case "txt":
		allProxies = loadFromTXT("proxies.txt")
	case "folder":
		allProxies = loadFromFolder("sources")
	default:
		fmt.Println("Unsupported mode.")
		return
	}

	seen := make(map[string]bool)
	var final []string
	for _, p := range allProxies {
		p = strings.TrimSpace(p)
		if strings.Contains(p, ":") && !seen[p] {
			seen[p] = true
			if onlyProtocol == "" || detectType(p) == onlyProtocol {
				final = append(final, p)
			}
		}
	}
		totalToCheck = len(final)
	if !silent {
		fmt.Printf("Checking %d proxies...\n", totalToCheck)
	}
	startTime = time.Now()
	sem := make(chan struct{}, 200)

	for _, proxy := range final {
		sem <- struct{}{}
		go check(proxy, detectType(proxy), sem)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	elapsed := time.Since(startTime).Seconds()
	if filesReady {
		writeSummary(elapsed)
		fmt.Println("\nResults saved in:", outputDir)
	}
	fmt.Printf("Done! Working: %d / %d | Time: %.2fs\n", successCount, totalToCheck, elapsed)
	if !silent {
		fmt.Print("Press ENTER to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func check(p, proto string, sem chan struct{}) {
	defer func() { <-sem }()
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	if proto == "http" {
		proxyURL, _ := url.Parse("http://" + p)
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	} else {
		dialer, err := proxy.SOCKS5("tcp", p, nil, proxy.Direct)
		if err != nil {
			increment()
			return
		}
		client.Transport = &http.Transport{Dial: dialer.Dial}
	}
	resp, err := client.Get(testURL)
	if err == nil && resp.StatusCode == 200 {
		mu.Lock()
		if !filesReady {
			outputDir = filepath.Join("results", time.Now().Format("2006-01-02_15-04-05"))
			os.MkdirAll(outputDir, 0755)
			httpFile, _ = os.Create(filepath.Join(outputDir, "http.txt"))
			socks4File, _ = os.Create(filepath.Join(outputDir, "socks4.txt"))
			socks5File, _ = os.Create(filepath.Join(outputDir, "socks5.txt"))
			summaryLog, _ = os.Create(filepath.Join(outputDir, "summary.log"))
			filesReady = true
		}
		successCount++
		writeProxy(p, proto)
		mu.Unlock()
	}
	increment()
}

func writeProxy(p, proto string) {
	var f *os.File
	switch proto {
	case "http":
		f = httpFile
	case "socks4":
		f = socks4File
	case "socks5":
		f = socks5File
	}
	if f != nil {
		f.WriteString(p + "\n")
	}
}

func detectType(p string) string {
	port := strings.Split(p, ":")
	if len(port) != 2 {
		return "http"
	}
	switch port[1] {
	case "1080":
		return "socks5"
	case "1081":
		return "socks4"
	default:
		return "http"
	}
}

func increment() {
	mu.Lock()
	totalChecked++
	elapsed := time.Since(startTime).Seconds()
	if !silent {
		speed := float64(totalChecked) / (elapsed + 0.1)
		fmt.Printf("\rProgress: %d/%d | ✅ %d | ⚡ %.1f req/s", totalChecked, totalToCheck, successCount, speed)
	}
	mu.Unlock()
}

func writeSummary(elapsed float64) {
	if summaryLog != nil {
		summaryLog.WriteString(fmt.Sprintf("Total: %d\nWorking: %d\nTime: %.2fs\n", totalToCheck, successCount, elapsed))
	}
}

func loadFromAPIs(path string) []string {
	var proxies []string
	file, err := os.Open(path)
	if err != nil {
		return proxies
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		link := strings.TrimSpace(scanner.Text())
		if link == "" || strings.HasPrefix(link, "#") {
			continue
		}
		found, err := fetchFromAPI(link)
		if err == nil {
			proxies = append(proxies, found...)
		}
	}
	return proxies
}

func loadFromTXT(path string) []string {
	var proxies []string
	file, err := os.Open(path)
	if err != nil {
		return proxies
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		proxies = append(proxies, line)
	}
	return proxies
}

func loadFromFolder(folder string) []string {
	var proxies []string
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".txt") {
			return nil
		}
		list := loadFromTXT(path)
		proxies = append(proxies, list...)
		return nil
	})
	return proxies
}

func fetchFromAPI(link string) ([]string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proxies []string
	ct := resp.Header.Get("Content-Type")

	if strings.Contains(ct, "application/json") {
		var result struct {
			Data []struct {
				IP   string `json:"ip"`
				Port string `json:"port"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			for _, item := range result.Data {
				if item.IP != "" && item.Port != "" {
					proxies = append(proxies, fmt.Sprintf("%s:%s", item.IP, item.Port))
				}
			}
		}
	} else if strings.Contains(ct, "text/html") {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}
		doc.Find("textarea").Each(func(i int, s *goquery.Selection) {
			lines := strings.Split(s.Text(), "\n")
			for _, l := range lines {
				if strings.Contains(l, ":") {
					proxies = append(proxies, strings.TrimSpace(l))
				}
			}
		})
		doc.Find("table").Each(func(i int, t *goquery.Selection) {
			t.Find("tr").Each(func(j int, row *goquery.Selection) {
				tds := row.Find("td")
				if tds.Length() >= 2 {
					ip := strings.TrimSpace(tds.Eq(0).Text())
					port := strings.TrimSpace(tds.Eq(1).Text())
					if strings.Contains(ip, ".") && port != "" {
						proxies = append(proxies, fmt.Sprintf("%s:%s", ip, port))
					}
				}
			})
		})
	} else {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, ":") {
				proxies = append(proxies, line)
			}
		}
	}
	return proxies, nil
}