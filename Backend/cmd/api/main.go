package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elvisouma/salmar-ai/internal/config"
	"github.com/elvisouma/salmar-ai/internal/handlers"
	"github.com/elvisouma/salmar-ai/internal/orchestrator"
	"github.com/elvisouma/salmar-ai/pkg/automation"
	"github.com/elvisouma/salmar-ai/pkg/client"
	"github.com/elvisouma/salmar-ai/pkg/ethics"
	"github.com/elvisouma/salmar-ai/pkg/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Initialize Python service client
	pythonClient := client.NewPythonServiceClient(cfg)

	// Initialize memory store
	memoryStore := memory.NewInMemoryStore()

	// Initialize ethical components
	contentFilter := ethics.NewContentFilter()
	reasoningTransparency := ethics.NewReasoningTransparency()

	// Initialize automation components
	taskScheduler := automation.NewTaskScheduler()
	integrationManager := automation.NewIntegrationManager()
	
	// Register default integrations
	integrationManager.RegisterIntegration(automation.NewWebhookIntegration(nil))
	integrationManager.RegisterIntegration(automation.NewEmailIntegration())
	
	// Create task handlers - commented out until needed
	// aiTaskHandler := automation.NewAITaskHandler(pythonClient)
	// integrationTaskHandler := automation.NewIntegrationTaskHandler(integrationManager)

	// Initialize orchestrator with all components
	orchService := orchestrator.New(cfg, pythonClient, memoryStore)
	
	// Set additional components in the orchestrator
	orchService.SetContentFilter(contentFilter)
	orchService.SetReasoningTransparency(reasoningTransparency)
	orchService.SetTaskScheduler(taskScheduler)
	orchService.SetIntegrationManager(integrationManager)

	// Register API routes
	handlers.RegisterRoutes(router, orchService)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start the task scheduler in a separate goroutine
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	defer schedulerCancel()
	
	if err := taskScheduler.Start(schedulerCtx); err != nil {
		log.Printf("Warning: Failed to start task scheduler: %v", err)
	} else {
		log.Println("Task scheduler started successfully")
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting Salmar AI server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop the task scheduler
	taskScheduler.Stop()
	log.Println("Task scheduler stopped")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}