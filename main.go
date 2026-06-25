package main

import (
"bufio"
"crypto/tls"
"fmt"
"io"
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


// 🛡️ Tự động vá lỗi
func autoRecover() {
if r := recover(); r != nil {
fmt.Printf("⚠️ Lỗi luồng: %v. Đang tự động bỏ qua...\n", r)
}
}

// 🎭 Fake IP trốn truy quét
func generateFakeIP() string {
return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(255), rand.Intn(255), rand.Intn(254)+1)
}

// 🔑 Cào Token từ Web
func scrapeToken() string {
if tokenURL == "" || tokenURL == "LINK_WEB_LAY_TOKEN_CUA_ONG" {
return "" // Nếu ông không điền thì bỏ qua bước này
}
fmt.Println("🔑 Đang cào Token mới nhất...")

customTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
client := &http.Client{Transport: customTransport, Timeout: 10 * time.Second}

req, _ := http.NewRequest("GET", tokenURL, nil)
req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Mobile Safari/537.36")

resp, err := client.Do(req)
if err != nil || resp.StatusCode != 200 {
fmt.Println("❌ Lỗi cào Token, dùng link gốc!")
return ""
}
defer resp.Body.Close()

bodyBytes, _ := io.ReadAll(resp.Body)
fmt.Println("✅ Đã lấy Token thành công!")
return strings.TrimSpace(string(bodyBytes)) // Trả về cục Token đã cào được
}

func extractUserAgent(vlcOpts []string) string {
for _, opt := range vlcOpts {
if strings.HasPrefix(opt, "#EXTVLCOPT:http-user-agent=") {
return strings.TrimPrefix(opt, "#EXTVLCOPT:http-user-agent=")
}
}
// Vũ khí mặc định của ông:
return "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Mobile Safari/537.36"
}

// 🚀 Kiểm tra link sống chết (Đã ép User Agent + Token)
func checkChannelHealth(url string, userAgent string) bool {
// Bỏ qua check link có PHP (vì nó thường tự redirect) hoặc link Clearkey DRM
if strings.Contains(url, ".php") || strings.Contains(url, ".mpd") {
return true 
}

customTransport := &http.Transport{
TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
MaxIdleConns:       100,
IdleConnTimeout:    90 * time.Second,
}

client := &http.Client{
Transport: customTransport,
Timeout:   6 * time.Second,
}

req, err := http.NewRequest("GET", url, nil)
if err != nil { return false }

req.Header.Set("User-Agent", userAgent)
req.Header.Set("Accept", "*/*")
req.Header.Set("X-Forwarded-For", generateFakeIP())
req.Header.Set("Client-IP", generateFakeIP())

resp, err := client.Do(req)
if err != nil { return false }
defer resp.Body.Close()

return resp.StatusCode == http.StatusOK
}

// 🚀 Cỗ máy chính
func runIptvScraperJob() {
defer autoRecover()

fmt.Printf("\n🔄 Kích hoạt Cỗ Máy (Token + Check Kênh)...\n")

// Lấy Token trước
token := scrapeToken()

inputFile, err := os.Open("source.m3u")
if err != nil {
fmt.Println("❌ Lỗi: Không thấy source.m3u (Ông dán m3u vào file này chưa?)")
return
}
defer inputFile.Close()

var channels []IPTVChannel
var currentChannel IPTVChannel
hasChannel := false

scanner := bufio.NewScanner(inputFile)
for scanner.Scan() {
line := strings.TrimSpace(scanner.Text())
if line == "" { continue }

if strings.HasPrefix(line, "#EXTINF:") {
if hasChannel { channels = append(channels, currentChannel) }
currentChannel = IPTVChannel{ExtInf: line}
hasChannel = true

idx := strings.LastIndex(line, ",")
if idx != -1 { currentChannel.Name = strings.TrimSpace(line[idx+1:]) } else { currentChannel.Name = "Kênh ẩn" }
} else if strings.HasPrefix(line, "#EXTVLCOPT:") {
currentChannel.VlcOpts = append(currentChannel.VlcOpts, line)
} else if strings.HasPrefix(line, "#KODIPROP:") {
currentChannel.KodiProps = append(currentChannel.KodiProps, line)
} else if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
url := line
// NẾU CÓ TOKEN, GHÉP VÀO LINK (Ông có thể tùy biến dòng này)
if token != "" && strings.Contains(url, ".m3u8") {
// VD: url = url + "?token=" + token
}
currentChannel.URL = url
}
}
if hasChannel { channels = append(channels, currentChannel) }

fmt.Printf("📡 Tìm thấy %d kênh. Bắt đầu ép xung đường truyền...\n", len(channels))

validChannels := make([]bool, len(channels))
var wg sync.WaitGroup
semaphore := make(chan struct{}, 15)

for i, ch := range channels {
if ch.URL == "" { continue }
wg.Add(1)
semaphore <- struct{}{}

go func(index int, c IPTVChannel) {
defer wg.Done()
defer func() { <-semaphore }()
defer autoRecover()

userAgent := extractUserAgent(c.VlcOpts)
if checkChannelHealth(c.URL, userAgent) {
validChannels[index] = true
fmt.Printf("✅ [LIVE] Xuyên thủng: %s\n", c.Name)
} else {
fmt.Printf("❌ [DIE] Bỏ qua: %s\n", c.Name)
}
}(i, ch)
}

wg.Wait()

outputFile, err := os.Create("live.m3u")
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

fmt.Printf("📊 Hoàn thành! Lọc được %d kênh sống. Lưu ra live.m3u.\n", successCount)

fmt.Println("☁️ Đẩy lên GitHub...")
exec.Command("git", "add", "live.m3u").Run()
exec.Command("git", "commit", "-m", "Auto update").Run()
exec.Command("git", "push", "origin", "main").Run()
}

func main() {
rand.Seed(time.Now().UnixNano())

// Chạy luôn lần 1
runIptvScraperJob()

// Lặp lại mỗi 1 tiếng
ticker := time.NewTicker(1 * time.Hour)
defer ticker.Stop()
for range ticker.C {
runIptvScraperJob()
}
}
