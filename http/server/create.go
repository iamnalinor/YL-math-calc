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
	"strconv"
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

	app := r.Context().Value("app").(*application.Application)

	app.Database.UpdatingMutex.Lock()

	opId, err := parseExpression(input.Expression, app)
	op, _ := app.Database.Get(opId)
	op.Expression = input.Expression
	app.Database.Update(op)

	app.Database.UpdatingMutex.Unlock()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to parse expression: %s", err)
		return

	}

	w.WriteHeader(http.StatusCreated)
	data, err := json.Marshal(createOutput{Id: opId})
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func parseExpression(expression string, app *application.Application) (operation.ID, error) {
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return 0, fmt.Errorf("failed to parse expression: %w", err)
	}

	result, err := unparseTree(expr, app)
	if err != nil {
		return 0, fmt.Errorf("failed to unparse expression: %w", err)
	}

	if result.OperationID == 0 {
		return 0, fmt.Errorf("the expression does not contain any operations")
	}

	return result.OperationID, nil
}

type unparseResult struct {
	OperationID operation.ID
	Value       float64
}

var unparseResultEmpty = unparseResult{}

func unparseTree(expr ast.Expr, app *application.Application) (unparseResult, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		left, err := unparseTree(e.X, app)
		if err != nil {
			return unparseResultEmpty, err
		}

		right, err := unparseTree(e.Y, app)
		if err != nil {
			return unparseResultEmpty, err
		}

		op := operation.Operation{}

		if left.OperationID != 0 {
			op.LeftOperationID = left.OperationID
		} else {
			op.Left = left.Value
		}

		if right.OperationID != 0 {
			op.RightOperationID = right.OperationID
		} else {
			op.Right = right.Value
		}

		switch e.Op {
		case token.ADD:
			op.Op = operation.Addition
		case token.SUB:
			op.Op = operation.Subtraction
		case token.MUL:
			op.Op = operation.Multiply
		case token.QUO:
			op.Op = operation.Division
		default:
			return unparseResultEmpty, fmt.Errorf("unsupported operation: %s", e.Op)
		}

		opID, err := app.Database.Create(op)

		if err != nil {
			return unparseResultEmpty, err
		}

		return unparseResult{OperationID: opID}, nil
	case *ast.BasicLit:
		value, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			return unparseResultEmpty, err
		}

		return unparseResult{Value: value}, nil
	case *ast.ParenExpr:
		return unparseTree(e.X, app)
	default:
		return unparseResultEmpty, fmt.Errorf("unsupported expression type: %T", e)
	}
}
