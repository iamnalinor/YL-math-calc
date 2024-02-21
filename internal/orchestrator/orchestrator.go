package orchestrator

import (
	"fmt"
	"math-calc/internal/application"
	"math-calc/internal/operation"
	"time"
)

type Orchestrator struct {
	app *application.Application
}

func New(app *application.Application) *Orchestrator {
	return &Orchestrator{
		app: app,
	}
}

// send with `go` statement is used to avoid deadlock
func send[T any](ch chan<- T, v T) {
	ch <- v
}

func (o *Orchestrator) Run() {
	// Starting workers

	workerIn := make(chan operation.ID) // Workers input channel
	orchIn := make(chan operation.ID)   // Orchestrator input channel, also workers output channel
	defer close(workerIn)
	defer close(orchIn)

	for i := 0; i < o.app.Config.GoroutineCount; i++ {
		go RunWorker(o.app, workerIn, orchIn)
	}

	go o.SearchOperations(orchIn)

	// Main cycle
	for id := range orchIn {
		o.app.Database.UpdatingMutex.Lock()

		op, _ := o.app.Database.Get(id)
		// Depending on the operation state, dealing with it
		switch op.State {
		case operation.StateCreated: // Sent from SearchOperations()
			fallthrough
		case operation.StateScheduled: // Sent from Run()
			if op.LeftOperationID != 0 || op.RightOperationID != 0 {
				op.State = operation.StateScheduled
				o.app.Database.Update(op)
				break
			}

			op.State = operation.StatePending
			o.app.Database.Update(op)
			fallthrough
		case operation.StatePending:
			workerIn <- id
			// State will be updated in RunWorker()
		case operation.StateProcessing:
			break
		case operation.StateDone: // Sent from RunWorker() and from itself
			allOps, _ := o.app.Database.All()
			for _, other := range allOps {
				if other.State != operation.StateScheduled {
					continue
				}

				if other.LeftOperationID == id {
					other.Left = op.Result
					other.LeftOperationID = 0
					o.app.Database.Update(other)
					go send(orchIn, other.Id)
				}
				if other.RightOperationID == id {
					other.Right = op.Result
					other.RightOperationID = 0
					o.app.Database.Update(other)
					go send(orchIn, other.Id)
				}
			}
		case operation.StateError: // Sent from RunWorker() and Run()
			allOps, _ := o.app.Database.All()
			for _, other := range allOps {
				if other.LeftOperationID == id || other.RightOperationID == id {
					other.Error = fmt.Errorf("sub-operation %d failed: %w", op.Id, op.Error)
					other.State = operation.StateError
					other.LeftOperationID = 0
					other.RightOperationID = 0
					o.app.Database.Update(other)
					go send(orchIn, other.Id)
				}
			}
		}

		o.app.Database.UpdatingMutex.Unlock()
	}
}

// SearchOperations periodically checks the database for operations of following states:
// - StateCreated
func (o *Orchestrator) SearchOperations(out chan<- operation.ID) {
	for {
		<-time.After(5 * time.Second)

		ops, err := o.app.Database.All()
		if err != nil {
			o.app.Logger.Printf("SearchOperations: failed to get operations: %s\n", err)
			continue
		}

		for id, op := range ops {
			if op.State == operation.StateCreated {
				o.app.Logger.Printf("SearchOperations: sent new operation %d to orchestrator\n", id)
				out <- id
			}
		}
	}
}
