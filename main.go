package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type IPTVChannel struct {
	ExtInf    string
	VlcOpts   []string
	KodiProps []string
	URL       string
	Name      string
}

// 🛡️ TÍNH NĂNG: Tự động sửa lỗi & Hồi phục hệ thống
func autoRecover() {
	if r := recover(); r != nil {
		fmt.Printf("⚠️ Hệ thống gặp nhiễu động: %v. Kích hoạt tự động hồi phục luồng và chạy tiếp!\n", r)
	}
}

// 🎭 TÍNH NĂNG: Giả lập IP (Tránh bị nhà đài quét Block IP Termux)
func generateFakeIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(255), rand.Intn(255), rand.Intn(254)+1)
}

func extractUserAgent(vlcOpts []string) string {
	for _, opt := range vlcOpts {
		if strings.HasPrefix(opt, "#EXTVLCOPT:http-user-agent=") {
			return strings.TrimPrefix(opt, "#EXTVLCOPT:http-user-agent=")
		}
	}
	// User-Agent dự phòng siêu mạnh
	return "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Mobile Safari/537.36"
}

// 🚀 TÍNH NĂNG: Xử lý & Bảo mật đường truyền
func checkChannelHealth(url string, userAgent string) bool {
	// Khởi tạo HTTP Transport bọc thép
	customTransport := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true}, // Bỏ qua lỗi chứng chỉ SSL rác của nhà mạng
		MaxIdleConns:       100,                                   // Tối ưu hóa giữ kết nối
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false, // Bật nén GZIP tăng tốc độ tải
	}

	client := &http.Client{
		Transport: customTransport,
		Timeout:   6 * time.Second, // Ép thời gian phản hồi, không để bot bị treo
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	// 🕵️ TÍNH NĂNG: Bơm Header giả lập hoàn hảo
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Forwarded-For", generateFakeIP()) // Fake IP Động
	req.Header.Set("Client-IP", generateFakeIP())

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func runIptvScraperJob() {
	defer autoRecover() // Bọc giáp chống sập cho toàn bộ chu kỳ

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\n🔄 [%s] Kích hoạt Cỗ Máy Quét (Chế độ: Đa luồng & Ẩn danh)...\n", currentTime)

	inputFile, err := os.Open("source.m3u")
	if err != nil {
		fmt.Printf("❌ Lỗi: Không tìm thấy source.m3u\n")
		return
	}
	defer inputFile.Close()

	var channels []IPTVChannel
	var currentChannel IPTVChannel
	hasChannel := false

	// Lọc file
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }

		if strings.HasPrefix(line, "#EXTINF:") {
			if hasChannel { channels = append(channels, currentChannel) }
			currentChannel = IPTVChannel{ExtInf: line}
			hasChannel = true

			idx := strings.LastIndex(line, ",")
			if idx != -1 {
				currentChannel.Name = strings.TrimSpace(line[idx+1:])
			} else {
				currentChannel.Name = "Kênh ẩn"
			}
		} else if strings.HasPrefix(line, "#EXTVLCOPT:") {
			currentChannel.VlcOpts = append(currentChannel.VlcOpts, line)
		} else if strings.HasPrefix(line, "#KODIPROP:") {
			currentChannel.KodiProps = append(currentChannel.KodiProps, line)
		} else if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			currentChannel.URL = line
		}
	}
	if hasChannel { channels = append(channels, currentChannel) }

	fmt.Printf("📡 Tìm thấy %d kênh. Bắt đầu ép xung đường truyền...\n", len(channels))

	// ⚡ TÍNH NĂNG: Tăng tốc đường truyền (15 luồng song song)
	validChannels := make([]bool, len(channels))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 15) // Giới hạn 15 kênh check cùng lúc

	for i, ch := range channels {
		if ch.URL == "" { continue }

		wg.Add(1)
		semaphore <- struct{}{} // Xếp hàng vào ống xả

		go func(index int, c IPTVChannel) {
			defer wg.Done()
			defer func() { <-semaphore }() // Trả chỗ khi chạy xong
			defer autoRecover()            // Chống sập cho từng luồng riêng lẻ

			userAgent := extractUserAgent(c.VlcOpts)
			if checkChannelHealth(c.URL, userAgent) {
				validChannels[index] = true
				fmt.Printf("✅ [LIVE] Xuyên thủng thành công: %s\n", c.Name)
			} else {
				fmt.Printf("❌ [DIE/BLOCKED] Bỏ qua: %s\n", c.Name)
			}
		}(i, ch)
	}

	// Đợi tất cả 15 luồng làm xong việc
	wg.Wait()

	// Ghi file theo đúng thứ tự gốc
	os.MkdirAll("data", os.ModePerm)
	outputFile, err := os.Create("data/live.m3u")
	if err != nil { return }
	defer outputFile.Close()

	outputFile.WriteString("#EXTM3U\n")
	successCount := 0

	for i, isValid := range validChannels {
		if isValid {
			successCount++
			ch := channels[i]
			outputFile.WriteString(ch.ExtInf + "\n")
			for _, opt := range ch.VlcOpts { outputFile.WriteString(opt + "\n") }
			for _, prop := range ch.KodiProps { outputFile.WriteString(prop + "\n") }
			outputFile.WriteString(ch.URL + "\n\n")
		}
	}

	fmt.Printf("📊 Hoàn thành! Bắt sống: %d kênh ổn định cao.\n", successCount)

	// Đẩy lên GitHub
	fmt.Println("🚀 Bảo mật kênh và đẩy lên đám mây GitHub...")
	exec.Command("bash", "./deploy.sh").Run()
	fmt.Println("🎉 Hoàn tất chu kỳ siêu tốc!")
}

func main() {
	// Cấp seed random cho trình tạo Fake IP
	rand.Seed(time.Now().UnixNano())

	// Chạy ngay lần đầu
	runIptvScraperJob()

	// Chu kỳ tuần hoàn 1 Tiếng / Lần
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	fmt.Println("\n🛡️ [HỆ THỐNG BỌC THÉP] Bot ngầm đang chạy (Bảo mật: Bật | Hồi phục: Bật).")

	for range ticker.C {
		runIptvScraperJob()
	}
}
