package routines

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"my-gocron/models"
	"time"
)

type CronService interface {
	StartJob(task models.Task) uuid.UUID
	ScheduleJobAt(task models.Task, at time.Time) uuid.UUID
	ScheduleJobEvery(task models.Task, interval time.Duration) uuid.UUID
	TryStopJob(workerID uuid.UUID)
	ListJobs() []uuid.UUID
}

type CronServiceImpl struct {
	workers map[uuid.UUID]*Worker
	// Unified error channel for all workers
	errChan chan RoutineError
	logger  slog.Logger
}

func NewCronService() CronService {
	master := &CronServiceImpl{
		workers: make(map[uuid.UUID]*Worker),
		errChan: make(chan RoutineError),
		logger:  *slog.Default(),
	}

	// Error handling routine
	go func() {
		for {
			select {
			case err := <-master.errChan:
				master.logger.Error("Error in worker:", "workerID", err.WorkerID, "error", err.Error())
				master.TryStopJob(err.WorkerID)
			}
		}
	}()

	// Worker cleanup routine
	go func() {
		for {
			for workerID, worker := range master.workers {
				select {
				case <-worker.finChan:
					delete(master.workers, workerID)
					master.logger.Info("Worker finished.", "workerID", workerID)
				default:
				}
			}
			// Sleep for 1 second
			<-time.After(time.Second)
		}
	}()

	return master
}

func (m *CronServiceImpl) StartJob(task models.Task) uuid.UUID {
	ctx, cancel := context.WithCancel(context.Background())
	worker := &Worker{
		ID:            uuid.New(),
		errChan:       &m.errChan,
		finChan:       make(chan bool),
		cancelContext: cancel,
	}
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
		default:
			// Execute the task
			err := task.Fn(ctx, task.Payload)
			if err != nil {
				routineErr := RoutineError{
					WorkerID: worker.ID,
					error:    err,
				}
				*worker.errChan <- routineErr
			}
		}
		// Signal the worker has finished so that it can be cleaned up
		worker.finChan <- true
	}(ctx)
	m.workers[worker.ID] = worker

	return worker.ID
}

func (m *CronServiceImpl) ScheduleJobAt(task models.Task, at time.Time) uuid.UUID {
	return m.StartJob(models.Task{
		Fn: func(ctx context.Context, payload any) error {
			for {
				select {
				case <-time.After(time.Until(at)):
					return task.Fn(ctx, task.Payload)
				case <-ctx.Done():
					return nil
				default:
					<-time.After(time.Second)
				}
			}
		},
		Payload: nil,
	})
}

func (m *CronServiceImpl) ScheduleJobEvery(task models.Task, interval time.Duration) uuid.UUID {
	return m.StartJob(models.Task{
		Fn: func(ctx context.Context, payload any) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					m.StartJob(task)
					<-time.After(interval)
				}
			}
		},
		Payload: nil,
	})
}

func (m *CronServiceImpl) TryStopJob(workerID uuid.UUID) {
	if worker, ok := m.workers[workerID]; ok {
		// Can only stop if the task implementation supports cancel signal
		worker.cancelContext()
	} else {
		m.logger.Error("Worker not found:", "workerID", workerID)
	}
}

func (m *CronServiceImpl) ListJobs() []uuid.UUID {
	workerIDs := make([]uuid.UUID, 0, len(m.workers))
	for workerID := range m.workers {
		workerIDs = append(workerIDs, workerID)
	}
	return workerIDs
}
