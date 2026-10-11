package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"backend/config"
	"backend/db"
	"backend/internal/infra/rabbitmq"
	"backend/internal/infra/rabbitmq/consumer"
	"backend/internal/infra/rabbitmq/producer"
	repoAI "backend/internal/repository/postgres/ai"
	"backend/internal/repository/redis"
	usecaseAI "backend/internal/usecase/ai"
	usecaseAPI "backend/internal/usecase/apicontract"
)

func main() {
	log.Println("==================================================")
	log.Println(" Flowak Asynchronous Background Worker Daemon     ")
	log.Println("==================================================")

	// 1. Initialize configuration
	config.InitConfig()

	// 2. Initialize PostgreSQL connection and schema
	db.InitDB()
	defer db.DB.Close()

	// 3. Initialize Redis infrastructure (with graceful degradation fallback)
	redisClient, err := redis.NewRedisClient(&config.ActiveConfig)
	if err != nil {
		log.Printf("[Worker] Warning: Redis client initialization skipped or offline: %v", err)
	} else if redisClient != nil {
		defer redisClient.Close()
		log.Println("[Worker] Redis connection established.")
	}

	// 4. Initialize RabbitMQ connection and channel
	rabbitConn, err := rabbitmq.NewRabbitMQConnection(config.ActiveConfig.RabbitMQURL)
	if err != nil {
		log.Fatalf("[Worker] Failed to connect to RabbitMQ broker: %v", err)
	}
	defer rabbitConn.Close()

	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("[Worker] Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	// 5. Ensure exchanges and queues exist
	if err := rabbitmq.SetupTopology(ch); err != nil {
		log.Fatalf("[Worker] Failed to declare RabbitMQ topology: %v", err)
	}

	// Fair dispatch: don't give more than 1 message to a worker at a time
	_ = ch.Qos(1, 0, false)

	aiRepo := repoAI.NewAIJobRepository(db.DB)

	// Setup context with graceful shutdown cancellation on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Listener 1: AI Generation & Audit Jobs
	go func() {
		defer wg.Done()
		log.Printf("[Worker] Listening on queue '%s' for AI workloads...", rabbitmq.QueueAIJobs)
		err := consumer.ConsumeAIJobs(ctx, ch, func(c context.Context, job producer.AIJobPayload) error {
			log.Printf("[Worker] Received AI job [id=%s, type=%s]", job.JobID, job.Type)
			_, jobErr := usecaseAI.ExecuteProcessAIFlowJob(c, aiRepo, job)
			if jobErr != nil {
				log.Printf("[Worker] AI job execution failed [id=%s]: %v", job.JobID, jobErr)
				return jobErr
			}
			log.Printf("[Worker] AI job completed successfully [id=%s]", job.JobID)
			return nil
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[Worker] AI jobs consumer stopped with error: %v", err)
		}
	}()

	// Listener 2: Server-Side API Contract Runner Jobs
	go func() {
		defer wg.Done()
		log.Printf("[Worker] Listening on queue '%s' for API runner jobs...", rabbitmq.QueueAPIRunnerJobs)
		err := consumer.ConsumeAPIRunnerJobs(ctx, ch, func(c context.Context, job producer.APIRunnerJobPayload) error {
			log.Printf("[Worker] Received API runner job [run_id=%s, request_id=%s]", job.RunID, job.APIRequestID)
			_, runErr := usecaseAPI.ExecuteProcessAPIRunJob(c, db.DB, job)
			if runErr != nil {
				log.Printf("[Worker] API runner job execution error [run_id=%s]: %v", job.RunID, runErr)
				// Job finished and recorded in DB; return nil so message is acknowledged
				return nil
			}
			log.Printf("[Worker] API runner job completed successfully [run_id=%s]", job.RunID)
			return nil
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[Worker] API runner consumer stopped with error: %v", err)
		}
	}()

	log.Println("[Worker] Background daemon running. Press Ctrl+C to terminate.")

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("[Worker] Termination signal received. Initiating graceful shutdown...")

	cancel()
	wg.Wait()
	log.Println("[Worker] All consumers stopped. Shutdown complete.")
}
