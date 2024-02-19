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
	"math-calc/internal/application"
	"math-calc/internal/operation"
	"net"
	"net/http"
)

var usedIdempotentTokens = make(map[string]bool)

type createInput struct {
	Expression string `json:"expression"`
}

type createOutput struct {
	Id operation.ID `json:"id"`
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

	app := r.Context().Value("app").(*application.Application)

	app.Database.UpdatingMutex.Lock()
	opMul, err := app.Database.Create(operation.Operation{
		Op:    operation.Multiply,
		Left:  2,
		Right: 2,
	})

	opAdd, err := app.Database.Create(operation.Operation{
		Op:               operation.Addition,
		Left:             2,
		RightOperationID: opMul,
		Expression:       "2+2*2",
	})
	app.Database.UpdatingMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	data, err := json.Marshal(createOutput{Id: opAdd})
	if err != nil {
		panic(err)
	}
	w.Write(data)
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
	app *application.Application,
) (func(context.Context) error, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/createExpression", createExpression)
	mux.HandleFunc("/expression/", getExpression)

	srv := &http.Server{
		Addr:    "localhost:8081",
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
