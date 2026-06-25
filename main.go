package main

import (
	"fmt"
	"os/exec"
	"time"
)

// Hàm thực hiện nhiệm vụ cào Link, Token và tự động Push
func runIptvScraperJob() {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\n🔄 [%s] Bắt đầu chu kỳ làm mới Token & Link...\n", currentTime)
	
	// TODO: Đoạn này chứa logic cào (scrape) của ông để xuất ra file new.m3u
	// Ví dụ: crawlDataToFile("new.m3u")
	fmt.Println("📝 Đã cập nhật xong dữ liệu mới vào file new.m3u cục bộ.")

	// Tự động gọi script deploy.sh để đẩy lên GitHub
	fmt.Println("🚀 Đang tự động kích hoạt deploy.sh để push lên GitHub...")
	cmd := exec.Command("bash", "./deploy.sh")
	
	// Chạy script và lấy kết quả trả về
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Lỗi tự động push: %v\n", err)
		fmt.Printf("Chi tiết lỗi:\n%s\n", string(output))
	} else {
		fmt.Println("🎉 Đã đồng bộ danh sách kênh mới lên GitHub thành công!")
	}
}

func main() {
	// Chạy ngay lập tức 1 lần khi vừa bật bot để cập nhật luôn
	runIptvScraperJob()

	// Thiết lập thời gian lặp lại: Đúng 1 tiếng (1 Hour)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	fmt.Println("\n📡 [STATUS] Bot đang chạy ngầm... Cứ đúng 1 tiếng sẽ tự động làm mới!")

	// Vòng lặp vô hạn chạy theo chu kỳ của ticker
	for range ticker.C {
		runIptvScraperJob()
	}
}
