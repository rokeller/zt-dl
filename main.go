package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rokeller/zt-dl/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		cancel() // Cancel the context on Ctrl+C
	}()

	cmd.Execute(ctx)
}
