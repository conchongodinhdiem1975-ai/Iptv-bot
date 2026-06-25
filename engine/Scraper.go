package engine

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// Channel định nghĩa cấu trúc của một kênh truyền hình
type Channel struct {
	Metadata string // Dòng #EXTINF...
	URL      string // Link stream
}

// Scraper service chịu trách nhiệm kiểm tra sức khỏe của các luồng
type Scraper struct {
	client *http.Client
	stats  *Stats // Giả lập theo dõi số liệu
}

type Stats struct {
	Success int
	Failed  int
	mu      sync.Mutex
}

// NewScraper khởi tạo Scraper với cấu hình tối ưu
func NewScraper(timeout time.Duration) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
			},
		},
		stats: &Stats{},
	}
}

// Execute thực thi việc cào và kiểm tra đồng loạt
func (s *Scraper) Execute(ctx context.Context, channels []Channel, workers int) []Channel {
	jobs := make(chan Channel, len(channels))
	results := make(chan Channel, len(channels))
	var wg sync.WaitGroup

	// Tạo pool các worker
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go s.worker(ctx, jobs, results, &wg)
	}

	// Đẩy job vào queue
	for _, ch := range channels {
		jobs <- ch
	}
	close(jobs)

	// Chờ đợi và đóng kênh kết quả
	go func() {
		wg.Wait()
		close(results)
	}()

	var finalChannels []Channel
	for ch := range results {
		finalChannels = append(finalChannels, ch)
	}

	log.Printf("📊 Scraper xong: Thành công: %d, Thất bại: %d\n", s.stats.Success, s.stats.Failed)
	return finalChannels
}

func (s *Scraper) worker(ctx context.Context, jobs <-chan Channel, results chan<- Channel, wg *sync.WaitGroup) {
	defer wg.Done()
	for ch := range jobs {
		if s.validate(ctx, ch.URL) {
			s.stats.incrementSuccess()
			results <- ch // Chỉ trả về kênh còn sống
		} else {
			s.stats.incrementFailed()
			log.Printf("❌ Kênh chết: %s", ch.Metadata)
		}
	}
}

// validate kiểm tra xem link còn sống không với cơ chế Retry
func (s *Scraper) validate(ctx context.Context, url string) bool {
	maxRetries := 2
	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

		resp, err := s.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond) // Đợi tí rồi retry
	}
	return false
}

func (s *Stats) incrementSuccess() {
	s.mu.Lock()
	s.Success++
	s.mu.Unlock()
}

func (s *Stats) incrementFailed() {
	s.mu.Lock()
	s.Failed++
	s.mu.Unlock()
}

