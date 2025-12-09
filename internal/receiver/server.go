package receiver

import (
	"context"
	"fmt"
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

// Server encapsulates the HTTP server and dependencies for the receiver.
type Server struct {
	cfg       *config.Config
	publisher *queue.ServiceBusPublisher
	httpServer *http.Server
}

// NewServer creates a new instance of the receiver server.
func NewServer(cfg *config.Config, publisher *queue.ServiceBusPublisher) *Server {
	s := &Server{
		cfg:       cfg,
		publisher: publisher,
	}
	// Setup the HTTP server with the handler logic
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.webhookHandler())
	mux.HandleFunc("/healthz", s.healthHandler())

	s.httpServer = &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}
	return s
}

// Run starts the HTTP server and manages graceful shutdown.
func (s *Server) Run(ctx context.Context) error {
	// Start the server in a goroutine
	go func() {
		log.Printf("Receiver Server listening on port %s...", s.cfg.Port)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("FATAL: HTTP server ListenAndServe: %v", err)
		}
	}()

	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Receiver Shutting down...")

	// Gracefully shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	log.Println("Receiver Server gracefully stopped.")
	return nil
}

// --- Handler Logic ---

func (s *Server) healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (s *Server) webhookHandler() http.HandlerFunc {
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
		if !github.ValidateSignature(signature, body, s.cfg.GitHubSecret) {
			log.Printf("Handler Error: Signature validation failed for IP %s", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// 3. LOGGING & IMMEDIATE RESPONSE
		event := r.Header.Get("X-GitHub-Event")
		log.Printf("Webhook Received: Event=%s, Length=%d bytes, Signature validated.", event, len(body))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Received and queueing for processing."))
		
		// 4. QUEUE ASYNCHRONOUSLY
		go func() {
			ctx := context.Background() 
			if err := s.publisher.Publish(ctx, body); err != nil {
				log.Printf("QUEUE FAILURE: Failed to publish validated payload (Event: %s): %v", event, err)
			} else {
				log.Printf("Queue Success: Payload successfully pushed (Event: %s).", event)
			}
		}()
	}
}
