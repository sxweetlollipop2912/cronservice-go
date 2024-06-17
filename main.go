package main

import (
	"context"
	"fmt"
	"my-gocron/models"
	"my-gocron/routines"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cronService := routines.NewCronService()

	cronService.ScheduleJobEvery(
		models.Task{
			Fn: func(ctx context.Context, a any) error {
				if aInt, ok := a.(int); ok {
					println("Hello", aInt)
				} else {
					return fmt.Errorf("payload is not an int")
				}
				return nil
			},
			Payload: 1,
		}, time.Second*2)

	cronService.ScheduleJobAt(
		models.Task{
			Fn: func(ctx context.Context, a any) error {
				if aStr, ok := a.(string); ok {
					println("Hello", aStr)
				} else {
					return fmt.Errorf("payload is not a string")
				}
				return nil
			},
			Payload: "world",
		}, time.Now().Add(time.Second*5))

	// Block so that the program doesn't exit.
	// This should have been done inside CronService.
	// But then we would need 2 servers, one for cron service and one for registering the tasks.
	// Still, this is just a temp solution.
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}
