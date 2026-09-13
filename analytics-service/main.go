package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	applicationClick "analytics-service/application/click"
	"analytics-service/container"
	messagequeue "analytics-service/message-queue"
)

func main() {

	app, err := container.NewContainer()
	if err != nil {
		log.Fatal("failed to initialize application:", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer app.Close()

	handler := func(payload []byte) error {
		if err := app.Click.UseCase.Handle(payload); err != nil {
			if errors.Is(err, applicationClick.ErrInvalidPayload) {
				log.Printf("dropping invalid click event: %v", err)
				return nil
			}
			return err
		}
		return nil
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		if err := app.Click.Consumer.Consume(ctx, handler); err != nil {
			if errors.Is(err, messagequeue.ErrQueueUnavailable) {
				log.Printf("consumer disabled: %v", err)
				return
			}
			log.Printf("consumer stopped: %v", err)
		}
	}()

	log.Println("click consumer running, waiting for shutdown signal")

	<-ctx.Done()
	log.Println("shutting down...")
	<-done
}