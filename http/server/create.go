package server

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"math-calc/internal/application"
	"math-calc/internal/operation"
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
