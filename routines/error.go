package routines

import "github.com/google/uuid"

type RoutineError struct {
	error
	WorkerID uuid.UUID
}
