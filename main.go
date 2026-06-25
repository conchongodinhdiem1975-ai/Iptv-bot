package main

import (
	"context"
	"iptv-bot/config"
	"iptv-bot/engine"
	"iptv-bot/storage"
	"log"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("🚀 Bot đã khởi động ở chế độ chạy nền (1 phút/lần)...")

	// Tạo ticker mỗi 1 phút
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Chạy lần đầu tiên ngay lập tức
	runJob(cfg)

	// Vòng lặp chờ ticker
	for range ticker.C {
		log.Println("🔄 Bắt đầu chu kỳ cào mới...")
		runJob(cfg)
	}
}

func runJob(cfg *config.Config) {
	scraper := engine.NewScraper(time.Duration(cfg.TimeoutSeconds) * time.Second)
	allChannels := storage.LoadPlaylist(cfg.InputFile)
	
	ctx := context.Background()
	liveChannels := scraper.Execute(ctx, allChannels, cfg.WorkerCount)
	
	storage.SavePlaylist(cfg.OutputFile, liveChannels)
	log.Println("✅ Hoàn tất chu kỳ. Đang đợi 1 phút nữa...")
}
