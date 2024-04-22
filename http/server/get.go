package server

import (
	"encoding/json"
	"fmt"
	"math-calc/internal/application"
	"math-calc/internal/operation"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type getResult struct {
	Id           operation.ID `json:"id"`
	Type         string       `json:"type"`
	Expression   string       `json:"expression"`
	Status       string       `json:"status"`
	Result       float64      `json:"result"`
	CreatedTime  time.Time    `json:"created_time"`
	FinishedTime time.Time    `json:"finished_time"`
}

func getExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "missing Authorization header")
		return
	}
	bearerToken, _ = strings.CutPrefix(bearerToken, "Bearer ")
	userId, err := checkJWT(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "invalid token")
		return
	}

	opIdRaw := r.URL.Path[len("/api/v1/expression/"):]
	if opIdRaw == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "operation id is not specified")
		return
	}

	opId, err := strconv.Atoi(opIdRaw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "operation id is not a number")
		return
	}

	app := r.Context().Value("app").(*application.Application)
	op, err := app.Database.Get(operation.ID(opId))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "operation not found")
		return
	}

	if op.OwnerID != userId {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "operation not found")
		return
	}

	status := ""
	switch op.State {
	case operation.StateCreated:
		status = "just created"
	case operation.StateScheduled:
		status = "waiting for other operation"
	case operation.StatePending:
		status = "in queue for calculation"
	case operation.StateProcessing:
		status = "calculating"
	case operation.StateDone:
		status = "done"
	case operation.StateError:
		status = fmt.Sprintf("error: %s", op.Error)
	}

	opType := "operation"
	if op.Expression != "" {
		opType = "expression"
	}

	result := getResult{
		Id:           op.Id,
		Type:         opType,
		Expression:   op.Expression,
		Status:       status,
		Result:       op.Result,
		CreatedTime:  op.CreatedTime,
		FinishedTime: op.FinishedTime,
	}
	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		app.Logger.Printf("failed to marshal result: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to marshal result")
		return
	}
	w.Write(data)
}
