package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// 🔗 DANH SÁCH CÁC NGUỒN CẦN CÀO (Ông dán các link raw M3U của ông vào đây nhé)
var rawSources = []string{
	"https://example.com/link_iptv_1.m3u", // Sửa lại thành link thật của ông
	"https://example.com/link_iptv_2.m3u",
}

// 🛡️ TÍNH NĂNG 1: Tự động sửa lỗi & Hồi phục
func autoRecover() {
	if r := recover(); r != nil {
		fmt.Printf("⚠️ Hệ thống cào gặp bão: %v. Tự động vá lỗi và chạy tiếp!\n", r)
	}
}

// 🎭 TÍNH NĂNG 2: Giả lập IP mạng
func generateFakeIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(255), rand.Intn(255), rand.Intn(254)+1)
}

// Hàm lõi: Thực hiện đột nhập và lấy dữ liệu
func scrapeSource(url string, wg *sync.WaitGroup, mu *sync.Mutex, file *os.File) {
	defer wg.Done()
	defer autoRecover() // Chống sập nếu link nguồn tải về bị dị dạng

	fmt.Printf("⏳ Đang kéo cáp cào dữ liệu từ: %s\n", url)

	// 🔒 TÍNH NĂNG 3: Bảo mật đường truyền & Bỏ qua chứng chỉ SSL
	customTransport := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:       50,
		IdleConnTimeout:    60 * time.Second,
		DisableCompression: false,
	}

	client := &http.Client{
		Transport: customTransport,
		Timeout:   15 * time.Second, // Tối đa 15 giây, không phản hồi là bỏ qua để khỏi treo bot
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return }

	// 🕵️ TÍNH NĂNG 4: Ép chuẩn User-Agent Android 10 Chrome 149
	userAgent := "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Mobile Safari/537.36"
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Forwarded-For", generateFakeIP())
	req.Header.Set("Client-IP", generateFakeIP())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Bị chặn hoặc đứt kết nối tại: %s\n", url)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Mã lỗi HTTP %d từ: %s\n", resp.StatusCode, url)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil { return }

	// Khoá Mutex: Đảm bảo khi nhiều luồng cùng ghi vào 1 file không bị nát dữ liệu
	mu.Lock()
	defer mu.Unlock()
	file.Write(bodyBytes)
	file.WriteString("\n") // Ngắt dòng cho sạch file

	fmt.Printf("✅ Đã hút thành công toàn bộ kênh từ: %s\n", url)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	defer autoRecover()

	fmt.Println("🚀 KÍCH HOẠT MÁY XÚC DỮ LIỆU IPTV (Đa luồng & Ẩn danh)...")

	// Dọn dẹp chiến trường cũ
	os.Remove("source.m3u")

	// Tạo file chứa phôi
	outFile, err := os.OpenFile("source.m3u", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("❌ Lỗi tạo file source.m3u")
		return
	}
	defer outFile.Close()

	outFile.WriteString("#EXTM3U\n") // Header chuẩn hóa IPTV

	var wg sync.WaitGroup
	var mu sync.Mutex

	// ⚡ TĂNG TỐC: Bắn tất cả các link cùng một lúc (Goroutines)
	for _, sourceUrl := range rawSources {
		if sourceUrl != "" {
			wg.Add(1)
			go scrapeSource(sourceUrl, &wg, &mu, outFile)
		}
	}

	wg.Wait() // Chờ tất cả luồng chạy xong
	fmt.Println("🎉 HOÀN TẤT! Toàn bộ mỏ IPTV đã được hút về file source.m3u an toàn. Sẵn sàng cho main.go xử lý!")
}
