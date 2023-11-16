// k8s: staging/src/k8s.io/apiserver/pkg/server/signal.go
// etcd: pkg/osutil/interrupt_unix.go
package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type InterruptHandler func()

var (
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
	shutdownHandler      chan os.Signal

	interruptRegisterMu sync.Mutex
	// interruptHandlers holds all registered InterruptHandlers in order
	// they will be executed.
	interruptHandlers = []InterruptHandler{}
)

// RegisterInterruptHandler registers a new InterruptHandler. Handlers registered
// after interrupt handing was initiated will not be executed.
func RegisterInterruptHandler(h InterruptHandler) {
	interruptRegisterMu.Lock()
	defer interruptRegisterMu.Unlock()
	interruptHandlers = append(interruptHandlers, h)
}

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
// Only one of SetupSignalContext and SetupSignalHandler should be called, and only can
// be called once.
func SetupSignalHandler() <-chan struct{} {
	return SetupSignalContext().Done()
}

// SetupSignalContext is same as SetupSignalHandler, but a context.Context is returned.
// Only one of SetupSignalContext and SetupSignalHandler should be called, and only can
// be called once.
func SetupSignalContext() context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	shutdownHandler = make(chan os.Signal, 2)

	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		<-shutdownHandler
		cancel()
		interruptRegisterMu.Lock()
		ihs := make([]InterruptHandler, len(interruptHandlers))
		copy(ihs, interruptHandlers)
		interruptRegisterMu.Unlock()
		for _, h := range ihs {
			h()
		}
		<-shutdownHandler
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}

// RequestShutdown emulates a received event that is considered as shutdown signal (SIGTERM/SIGINT)
// This returns whether a handler was notified
func RequestShutdown() bool {
	if shutdownHandler != nil {
		select {
		case shutdownHandler <- shutdownSignals[0]:
			return true
		default:
		}
	}

	return false
}
