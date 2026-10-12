package main

import (
	"context"
	"fmt"
	"migrations/internal"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	args, err := internal.ParseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	cfg, err := internal.LoadConfig("config.yml")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	return internal.Execute(ctx, args, cfg)
}
