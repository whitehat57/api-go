package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type WordPressAPITester struct {
	UserAgents []string
	Headers    map[string]string
	Endpoints  []string
	Payloads   map[string]map[string]string
}

func NewTester() *WordPressAPITester {
	return &WordPressAPITester{
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			"Mozilla/5.0 (X11; Linux x86_64) Chrome/91.0.4472.124 Safari/537.36",
		},
		Headers: map[string]string{
			"Accept":           "application/json, text/plain, */*",
			"Accept-Language": "en-US,en;q=0.9",
			"Connection":       "keep-alive",
			"Content-Type":     "application/json",
			"X-Requested-With": "XMLHttpRequest",
		},
		Endpoints: []string{
			"wp-json/wp/v2/posts", "wp-json/wp/v2/users", "wp-json/wp/v2/comments",
			"wp-json", "wp-admin", "wp-login.php", "xmlrpc.php", "wp-content",
		},
		Payloads: map[string]map[string]string{
			"auth_test": {"username": "admin", "password": "admin123"},
			"post_test": {"title": "Spam", "content": "Spam Content", "status": "publish"},
		},
	}
}

func (tester *WordPressAPITester) GetRandomUserAgent() string {
	return tester.UserAgents[rand.Intn(len(tester.UserAgents))]
}

func (tester *WordPressAPITester) SendRequest(method, url, endpoint string, data map[string]string) {
	fullURL := fmt.Sprintf("%s/%s", strings.TrimRight(url, "/"), strings.TrimLeft(endpoint, "/"))
	req, _ := http.NewRequest(method, fullURL, nil)

	if method == "POST" {
		body, _ := json.Marshal(data)
		req, _ = http.NewRequest(method, fullURL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("User-Agent", tester.GetRandomUserAgent())
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] %s %s: %v\n", method, fullURL, err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("[%d] %s %s: %s\n", resp.StatusCode, method, fullURL, string(body))
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Masukkan URL target (contoh: https://example.com): ")
	targetURL, _ := reader.ReadString('\n')
	targetURL = strings.TrimSpace(targetURL)

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		fmt.Println("[ERROR] URL tidak valid. Harap sertakan http:// atau https://")
		return
	}

	tester := NewTester()

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				endpoint := tester.Endpoints[rand.Intn(len(tester.Endpoints))]
				tester.SendRequest("GET", targetURL, endpoint, nil)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(300)))
			}
		}()
	}
	wg.Wait()
}
