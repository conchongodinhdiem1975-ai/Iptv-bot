# 📺 IPTV Bot Pro

[![Status](https://img.shields.io/badge/status-active-brightgreen.svg)]() 
[![Language](https://img.shields.io/badge/language-Go-blue.svg)]()

Chào mừng đến với **IPTV Bot** – Công cụ tự động hóa việc thu thập, kiểm tra và quản lý danh sách kênh truyền hình (M3U) cá nhân.

## 🚀 Tính năng nổi bật
* **Tự động hóa:** Tự động cào dữ liệu, lọc kênh "chết" và cập nhật danh sách mới.
* **Tốc độ cao:** Được viết bằng Go (Golang) cho hiệu năng cực nhanh.
* **Tương thích:** Tối ưu hóa để chạy mượt mà trên **Termux** (Android).
* **Quản lý thông minh:** Kiểm tra trạng thái stream theo thời gian thực.

## 🛠 Cách cài đặt
Để chạy con bot này, hãy đảm bảo ông đã cài đặt Go và Git trên Termux:

```bash
git clone [https://github.com/conchongodinhdiem1975-ai/Iptv-bot.git](https://github.com/conchongodinhdiem1975-ai/Iptv-bot.git)
cd Iptv-bot
go build -o iptv-bot
./iptv-bot

