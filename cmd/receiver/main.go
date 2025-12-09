package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webhook-handler-project/internal/config"
	"webhook-handler-project/internal/github"
	"webhook-handler-project/internal/queue"
)

func main() {
	// 1. Load Configuration (only receiver-relevant fields need to be present)
	cfg := config.LoadBaseConfig()
    if cfg.GitHubSecret == "" {
        log.Fatal("FATAL: GITHUB_WEBHOOK_SECRET is required for the receiver.")
    }

	// 2. Initialize Azure Service Bus Publisher
	publisher, err := queue.NewServiceBusPublisher(cfg.ServiceBusConnectionString, cfg.ServiceBusQueueName)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Service Bus publisher: %v", err)
	}
	defer publisher.Close(context.Background())

	// 3. Setup HTTP Router
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler(cfg, publisher))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 4. Start HTTP Server and Graceful Shutdown
	server := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	// ... (Graceful shutdown logic from previous response)
    
	go func() {
		log.Printf("Receiver Server listening on port %s...", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("FATAL: HTTP server ListenAndServe: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop 

	log.Println("Receiver Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server shutdown failed: %v", err)
	}
	log.Println("Receiver Server gracefully stopped.")
}

func webhookHandler(cfg *config.Config, publisher *queue.ServiceBusPublisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// 1. READ RAW BODY
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Handler Error: Failed to read request body: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// 2. VALIDATE SIGNATURE
		signature := r.Header.Get("X-Hub-Signature-256")
		if !github.ValidateSignature(signature, body, cfg.GitHubSecret) {
			log.Printf("Handler Error: Signature validation failed for IP %s", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// 3. LOGGING
		event := r.Header.Get("X-GitHub-Event")
		log.Printf("Webhook Received: Event=%s, Length=%d bytes, Signature validated.", event, len(body))


		// 4. IMMEDIATE RESPONSE TO GITHUB (HTTP 200 OK)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Received and queueing for processing."))
		
		// 5. QUEUE THE RAW PAYLOAD ASYNCHRONOUSLY
		go func() {
			ctx := context.Background() 
			if err := publisher.Publish(ctx, body); err != nil {
				log.Printf("QUEUE FAILURE: Failed to publish validated payload (Event: %s): %v", event, err)
			} else {
				log.Printf("Queue Success: Payload successfully pushed (Event: %s).", event)
			}
		}()
	}
}
