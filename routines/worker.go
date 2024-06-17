package routines

import (
	"context"
	"github.com/google/uuid"
)

type Worker struct {
	ID            uuid.UUID
	errChan       *chan RoutineError
	finChan       chan bool
	cancelContext context.CancelFunc
}
