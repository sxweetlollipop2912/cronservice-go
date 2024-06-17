package models

import "context"

type Task struct {
	Fn      func(context.Context, any) error
	Payload any
}
