package async

import (
	"context"
	"runtime/debug"

	"github.com/reaburoa/micro-kit/utils/log"
)

func RunWithContext(ctx context.Context, f func() error) error {
	errChan := make(chan error)
	go func() {
		errChan <- f()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func RunWithRecover(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic: %v", r)
			}
			stackList := string(debug.Stack())
			log.Errorf("Stack trace:\n %s", stackList)
		}()
		f()
	}()
}
