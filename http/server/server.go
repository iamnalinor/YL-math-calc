package server

import (
	"context"
	"log"
	"math-calc/internal/application"
	"net"
	"net/http"
)

func Run(
	app *application.Application,
) (func(context.Context) error, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", userRegister)
	mux.HandleFunc("/api/v1/login", userLogin)
	mux.HandleFunc("/api/v1/createExpression", createExpression)
	mux.HandleFunc("/api/v1/expression/", getExpression)

	srv := &http.Server{
		Addr:    "0.0.0.0:8081",
		Handler: loggingMiddleware(app.Logger)(mux),
		BaseContext: func(listener net.Listener) context.Context {
			return context.WithValue(context.Background(), "app", app)
		}}

	go func() {
		// Запускаем сервер
		if err := srv.ListenAndServe(); err != nil {
			app.Logger.Fatal("ListenAndServe", err)
		}
	}()
	// вернем функцию для завершения работы сервера
	return srv.Shutdown, nil
}

// middleware для логированя запросов
func loggingMiddleware(logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("%s %s\n",
				r.Method,
				r.URL.Path,
			)

			next.ServeHTTP(w, r)
		})
	}
}
