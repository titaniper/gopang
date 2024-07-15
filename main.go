package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

// setJSONContentType는 모든 응답에 Content-Type 헤더를 설정하는 미들웨어입니다.
func setJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

const maxConcurrentRequests = 3

var mu sync.Mutex
var currentRequests int
var count int

func worker(txId string, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	// 임계 구역에 들어가기 전에 잠금
	mu.Lock()
	if currentRequests >= maxConcurrentRequests {
		fmt.Printf("Return %s %d, %d, %d\n", txId, id, count, currentRequests)
		mu.Unlock() // 조건이 만족되지 않으면 잠금을 해제하고 리턴
		return
	}
	fmt.Printf("gogo1 %s %d, %d, %d\n", txId, id, count, currentRequests)
	currentRequests++
	count++
	mu.Unlock()

	// 작업 수행
	fmt.Printf("Worker %s %d is doing work\n", txId, id)

	fmt.Printf("gogo2 %s %d, %d, %d\n", txId, id, count, currentRequests)
	time.Sleep(1 * time.Second) // 작업을 시뮬레이션

	fmt.Printf("gogo3 %s %d, %d, %d\n", txId, id, count, currentRequests)

	// 작업 완료 후 임계 구역에 들어가기 전에 잠금
	mu.Lock()
	fmt.Printf("gogo4 %s %d, %d, %d\n", txId, id, count, currentRequests)
	currentRequests--
	mu.Unlock()
}

func writeResponse(w http.ResponseWriter, count int) {
	json.NewEncoder(w).Encode(map[string][]string{
		"data": {"하이", fmt.Sprintf("하이 %d", count)},
	})
}

func clientsHandler(w http.ResponseWriter, r *http.Request) {
	// 새로운 UUID 생성
	txId := uuid.New().String()
	fmt.Printf("Generated txId: %s\n", txId)

	//if r.Method == http.MethodGet {
	//	w.Header().Set("Content-Type", "application/json")
	//	defer writeResponse(w, count)
	//	var wg sync.WaitGroup
	//	for i := 0; i < 10; i++ {
	//		wg.Add(1)
	//		go worker(txId, i, &wg)
	//		fmt.Printf("count: %s, %d, %d, %d\n", txId, i, count, currentRequests)
	//	}
	//	wg.Wait()
	//} else {
	//	w.WriteHeader(http.StatusMethodNotAllowed)
	//	fmt.Fprintf(w, "405 - Method Not Allowed")
	//}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		defer writeResponse(w, count)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "405 - Method Not Allowed")
	}
}

// notFoundHandler는 허용되지 않은 경로에 대한 핸들러입니다.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "404 - Not Found")
}

func main() {
	// New Server Multiplexer
	// 네트워크와 전자공학에서 멀티플렉서는 여러 신호를 하나의 신호로 결합하는 장치입니다. 반대로 디멀티플렉서(Demultiplexer)는 하나의 신호를 여러 신호로 분리합니다.
	mux := http.NewServeMux()

	// 기본 핸들러 외의 모든 경로에 대해 notFoundHandler 설정
	mux.HandleFunc("/", notFoundHandler)
	mux.HandleFunc("/admins/v1/clients", clientsHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 미들웨어를 사용하여 Content-Type 설정
	wrappedMux := setJSONContentType(mux)

	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		fmt.Println(err)
	}
}
