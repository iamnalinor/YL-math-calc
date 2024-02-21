package orchestrator

import (
	"fmt"
	"math-calc/internal/application"
	"math-calc/internal/operation"
	"time"
)

func RunWorker(app *application.Application, in <-chan operation.ID, out chan<- operation.ID) {
	for op := range in {
		app.Database.UpdatingMutex.Lock()
		op, _ := app.Database.Get(op)
		op.State = operation.StateProcessing
		app.Database.Update(op)
		app.Database.UpdatingMutex.Unlock()

		app.Logger.Printf("worker: operation%d: started\n", op.Id)

		duration := time.Duration(app.Config.OperationCalculationTime) * time.Second
		result, err := Calculate(op.Op, op.Left, op.Right, duration)
		op, _ = app.Database.Get(op.Id)

		app.Database.UpdatingMutex.Lock()
		if err != nil {
			op.State = operation.StateError
			op.Error = fmt.Errorf("calculate failed: %w", err)
			app.Logger.Printf("worker: operation%d: failed to calculate: %s\n", op.Id, err)
		} else {
			op.State = operation.StateDone
			op.Result = result
			app.Logger.Printf("worker: operation%d: calculated successfully, result is %f\n", op.Id, result)
		}

		op.FinishedTime = time.Now()

		app.Database.Update(op)
		app.Database.UpdatingMutex.Unlock()

		out <- op.Id
	}
}

func Calculate(op operation.Operator, left, right float64, duration time.Duration) (float64, error) {
	// Implement some delay to simulate real work
	<-time.After(duration)

	switch op {
	case operation.Addition:
		return left + right, nil
	case operation.Subtraction:
		return left - right, nil
	case operation.Multiply:
		return left * right, nil
	case operation.Division:
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	}
	return 0, fmt.Errorf("unknown operation %s", op)
}
