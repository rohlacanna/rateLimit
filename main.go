package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	requestsPerIP     = 10
	rateLimitRequests = 3
	rateLimitPerSec   = 1
	serverPort        = ":8080"
)

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

type RequestResult struct {
	IP        string
	ReqNum    int
	Status    int
	Duration  time.Duration
	Error     error
	Timestamp time.Time
}

func getIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rateLimitPerSec, rateLimitRequests)
		visitors[ip] = limiter
	}

	return limiter
}

func rateLimitByIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := getVisitor(ip)
		if !limiter.Allow() {
			fmt.Printf("Request from IP %s was denied\n", ip)

			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		fmt.Printf("Request from IP %s was allowed\n", ip)

		next.ServeHTTP(w, r)
	})
}

func makeRequest(ip string, reqNum int) RequestResult {
	startTime := time.Now()

	req, _ := http.NewRequest("GET", "http://localhost"+serverPort, nil)
	req.RemoteAddr = ip + ":123456"
	resp, err := http.DefaultClient.Do(req)

	result := RequestResult{
		IP:        ip,
		ReqNum:    reqNum,
		Duration:  time.Since(startTime),
		Timestamp: startTime,
		Error:     err,
	}

	if err == nil {
		result.Status = resp.StatusCode
		resp.Body.Close()
	}

	return result
}

func formatResult(result RequestResult) string {
	if result.Error != nil {
		return fmt.Sprintf("[%s] IP: %s | Req #%d | ERROR: %v | Duration: %v",
			result.Timestamp.Format("15:04:05.000"),
			result.IP,
			result.ReqNum,
			result.Error,
			result.Duration)
	}

	status := "❌"
	if result.Status == 200 {
		status = "✅"
	}

	return fmt.Sprintf("[%s] IP: %s | Req #%d | Status: %d %s | Duration: %.2fms",
		result.Timestamp.Format("15:04:05.000"),
		result.IP,
		result.ReqNum,
		result.Status,
		status,
		float64(result.Duration.Microseconds())/1000.0)
}

func doRequestUsingDifferentIP() {
	fmt.Println("=== Starting tests with different IPs ===")
	ips := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"}

	results := make(chan string, len(ips)*requestsPerIP)
	var wg sync.WaitGroup

	for _, ip := range ips {
		for i := 0; i < requestsPerIP; i++ {
			wg.Add(1)
			go func(ip string, reqNum int) {
				defer wg.Done()
				result := makeRequest(ip, reqNum)
				results <- formatResult(result)
			}(ip, i+1)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("\nRequest Results:")
	fmt.Println("----------------------------------------")

	ipResults := make(map[string][]string)
	successCount := make(map[string]int)

	for result := range results {
		for _, ip := range ips {
			if strings.Contains(result, ip) {
				ipResults[ip] = append(ipResults[ip], result)
				if strings.Contains(result, "Status: 200") {
					successCount[ip]++
				}
			}
		}
	}

	printResults(ipResults, successCount, ips)
}

func printResults(ipResults map[string][]string, successCount map[string]int, ips []string) {
	for _, ip := range ips {
		fmt.Printf("\nResults for IP %s:\n", ip)
		fmt.Println("----------------------------------------")
		for _, result := range ipResults[ip] {
			fmt.Println(result)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("----------------------------------------")
	for _, ip := range ips {
		fmt.Printf("IP %s: %d successful / %d total requests (Rate limit: %d requests)\n",
			ip, successCount[ip], requestsPerIP, rateLimitRequests)
	}
	fmt.Printf("\nTotal requests processed: %d\n", len(ips)*requestsPerIP)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))

	go http.ListenAndServe(":8080", rateLimitByIP(mux))

	doRequestUsingDifferentIP()
}
