package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"iptv-bot/engine" // Import để dùng struct Channel của engine
)

// LoadPlaylist đọc file m3u và tách Metadata + URL
func LoadPlaylist(filename string) []engine.Channel {
	var channels []engine.Channel
	
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("❌ Không thể mở file: %v\n", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastExtInf string
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, "#EXTINF") {
			lastExtInf = line
		} else if !strings.HasPrefix(line, "#") && line != "" && lastExtInf != "" {
			// Đây là URL hợp lệ, ghép với EXTINF trước đó
			channels = append(channels, engine.Channel{
				Metadata: lastExtInf,
				URL:      line,
			})
			lastExtInf = "" // Reset sau khi đã ghép
		}
	}
	return channels
}

// SavePlaylist ghi file kết quả với định dạng chuẩn M3U
func SavePlaylist(filename string, channels []engine.Channel) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Lỗi ghi file: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString("#EXTM3U\n")
	for _, ch := range channels {
		file.WriteString(ch.Metadata + "\n")
		file.WriteString(ch.URL + "\n")
	}
	fmt.Printf("✅ Đã lưu xong %d kênh vào file: %s\n", len(channels), filename)
}

