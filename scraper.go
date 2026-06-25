package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// 🔗 1. CHỖ NÀY ÔNG DÁN MẤY LINK PLAYLIST BUỔI CHIỀU VÀO NHÉ
var rawSources = []string{
	"LINK_PLAYLIST_CUA_ONG_1", 
	"LINK_PLAYLIST_CUA_ONG_2",
}

// 🔗 2. CHỖ NÀY DÀNH CHO LINK CÀO TOKEN CỦA ÔNG
var tokenURL = "LINK_WEB_LAY_TOKEN_CUA_ONG"

// 🛡️ Tự động vá lỗi (Tránh sập bot)
func autoRecover() {
	if r := recover(); r != nil {
		fmt.Printf("⚠️ Lỗi ngoại lệ: %v. Đang tự động hồi phục...\n", r)
	}
}

// 🎭 Fake IP trốn truy quét
func generateFakeIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(255), rand.Intn(255), rand.Intn(254)+1)
}

// 🕵️ Hàm khởi tạo Request bọc thép (Xài User-Agent Android 10 ông cấp)
func createSecureRequest(url string) (*http.Client, *http.Request) {
	customTransport := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:       50,
		IdleConnTimeout:    60 * time.Second,
	}

	client := &http.Client{
		Transport: customTransport,
		Timeout:   15 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)
	
	// Ép chuẩn User-Agent Android 10 Chrome 149
	userAgent := "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Mobile Safari/537.36"
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Forwarded-For", generateFakeIP())
	req.Header.Set("Client-IP", generateFakeIP())

	return client, req
}

// 🔑 Cào Token
func scrapeToken() string {
	if tokenURL == "" || tokenURL == "LINK_WEB_LAY_TOKEN_CUA_ONG" {
		return ""
	}
	fmt.Println("🔑 Đang cào Token mới nhất...")
	client, req := createSecureRequest(tokenURL)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("❌ Lỗi cào Token!")
		return ""
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	// (Lưu ý: Nếu Token của ông nằm trong JSON hoặc cần lọc regex, ông có thể bổ sung mã xử lý chuỗi ở đây)
	fmt.Println("✅ Đã lấy Token thành công!")
	return string(bodyBytes)
}

// 📡 Cào Playlist
func scrapeSource(url string, wg *sync.WaitGroup, mu *sync.Mutex, file *os.File, token string) {
	defer wg.Done()
	defer autoRecover()

	fmt.Printf("⏳ Đang xúc m3u từ: %s\n", url)
	client, req := createSecureRequest(url)
	
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Chặn/Lỗi tại: %s\n", url)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil { return }

	// Xử lý chèn Token vào kênh nếu cần thiết (Tùy thuộc cấu trúc kênh của ông)
	// content := strings.ReplaceAll(string(bodyBytes), "TOKEN_CU", token)

	mu.Lock()
	defer mu.Unlock()
	file.Write(bodyBytes) // Đổ thẳng vào file live.m3u
	file.WriteString("\n")

	fmt.Printf("✅ Xúc xong: %s\n", url)
}

// 🚀 Đẩy tự động lên GitHub
func pushToGitHub() {
	fmt.Println("☁️ Đang commit và đẩy live.m3u lên GitHub...")
	
	cmd1 := exec.Command("git", "add", "live.m3u")
	cmd1.Run()

	commitMsg := fmt.Sprintf("Auto-update channels & tokens: %s", time.Now().Format("2006-01-02 15:04:05"))
	cmd2 := exec.Command("git", "commit", "-m", commitMsg)
	cmd2.Run()

	cmd3 := exec.Command("git", "push", "origin", "main") // Đổi 'main' thành 'master' nếu repo của ông xài master
	err := cmd3.Run()

	if err != nil {
		fmt.Println("❌ Lỗi đẩy lên GitHub! Chắc chưa thiết lập SSH/Token Gihub trên máy.")
	} else {
		fmt.Println("🎉 Đã đẩy thành công lên kho chứa đám mây!")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	defer autoRecover()

	fmt.Println("🚀 KÍCH HOẠT MÁY XÚC (Cào Playlist + Token -> live.m3u -> GitHub)")

	// Lấy Token trước
	token := scrapeToken()

	// Dọn file cũ
	os.Remove("live.m3u")
	outFile, err := os.OpenFile("live.m3u", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("❌ Lỗi tạo live.m3u")
		return
	}
	defer outFile.Close()

	outFile.WriteString("#EXTM3U\n")

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Cào đa luồng
	for _, sourceUrl := range rawSources {
		if sourceUrl != "" && sourceUrl != "LINK_PLAYLIST_CUA_ONG_1" {
			wg.Add(1)
			go scrapeSource(sourceUrl, &wg, &mu, outFile, token)
		}
	}
	wg.Wait()
	
	// Gọi hàm xả thẳng lên Github
	pushToGitHub()
}
