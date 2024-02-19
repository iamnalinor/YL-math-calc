package operation

import "time"

// ID is a unique identifier of an operation.
type ID int64

// Operator is a type of binary operation represented by string.
type Operator string
type State uint8

const (
	Addition    Operator = "+"
	Subtraction Operator = "-"
	Multiply    Operator = "*"
	Division    Operator = "/"
)

const (
	// StateCreated represents a default value of State.
	// Orchestrator should change the state to either StateScheduled or StatePending immediately.
	StateCreated State = iota
	// StateScheduled means that some other operations have to be completed to start this one.
	// In this case, LeftOperationID or RightOperationID is not empty.
	StateScheduled
	// StatePending means that the operation is ready to be processed by a worker.
	StatePending
	// StateProcessing means that the operation is currently running on a worker.
	StateProcessing
	// StateDone means that the operation has been completed successfully.
	// In this case, Result is not empty.
	StateDone
	// StateError means that either this operation, LeftOperationID or RightOperationID has failed.
	// In this case, Error is not empty.
	StateError
)

type Operation struct {
	Id ID
	// Op represents type of operation that is performed on left and right values.
	Op          Operator
	State       State
	CreatedTime time.Time
	// FinishedTime is empty while State is not StateDone or StateError.
	FinishedTime time.Time

	// Left is the value of left operand. It can be empty if LeftOperationID is set.
	Left float64
	// Right is the value of right operand. It can be empty if RightOperationID is set.
	Right float64

	// LeftOperationID is the ID of the operation which result is used as left operand.
	LeftOperationID ID
	// RightOperationID is the ID of the operation which result is used as right operand.
	RightOperationID ID

	Result float64
	Error  error

	// Expression field can be set in order to store the original expression.
	// This doesn't influence orchestrator and workers in any way.
	Expression string
}
