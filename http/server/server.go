package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

var usedIdempotentTokens = make(map[string]bool)

func create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
	usedIdempotentTokens[r.Header.Get("Idempotentcy-Token")] = true
}

func getExpression(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func Run(
	ctx context.Context,
	logger *log.Logger,
) (func(context.Context) error, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/create", create)
	mux.HandleFunc("/get", getExpression)

	srv := &http.Server{Addr: ":8081", Handler: loggingMiddleware(logger)(mux)}

	go func() {
		// Запускаем сервер
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal("ListenAndServe", err)
		}
	}()
	// вернем функцию для завершения работы сервера
	return srv.Shutdown, nil
}

// middleware для логированя запросов
func loggingMiddleware(logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Пропуск запроса к следующему обработчику
			next.ServeHTTP(w, r)

			// Завершение логирования после выполнения запроса
			duration := time.Since(start)
			logger.Printf("HTTP request %s, %s, %d sec\n",
				r.Method,
				r.URL.Path,
				duration/time.Second,
			)
		})
	}
}
