package server

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"net/http"
)

var usedIdempotentTokens = make(map[string]bool)

type createInput struct {
	Expression string `json:"expression"`
}

func createExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "only POST requests are allowed")
		return
	}

	idempToken := r.Header.Get("X-Idempotency-Token")
	if idempToken != "" {
		if usedIdempotentTokens[idempToken] {
			fmt.Fprintln(w, "Idempotency token is already used")
			return
		}

		usedIdempotentTokens[idempToken] = true
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to read request body: %s", err)
		return
	}

	input := createInput{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to unparse json: %s", err)
		return
	}

	parseExpression(input.Expression)

	w.Write([]byte("Hello, World!"))
}

func parseExpression(expression string) (string, error) {
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return "", fmt.Errorf("failed to parse expression: %w", err)
	}

	ast.Print(token.NewFileSet(), expr)
	return "", nil
}

func getExpression(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func Run(
	logger *log.Logger,
) (func(context.Context) error, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/createExpression", createExpression)
	mux.HandleFunc("/expression/", getExpression)

	srv := &http.Server{Addr: "localhost:8081", Handler: loggingMiddleware(logger)(mux)}

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
			logger.Printf("%s %s\n",
				r.Method,
				r.URL.Path,
			)

			next.ServeHTTP(w, r)
		})
	}
}
