package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"

	"twilio-go/internal/handler"
	"twilio-go/internal/repository"
	"twilio-go/internal/service"
	"twilio-go/internal/telephony"
)

func main() {
	ctx := context.Background()

	firestoreClient, err := firestore.NewClient(
		ctx,
		"twilio-go-5b129",
	)

	if err != nil {
		log.Fatal(err)
	}

	defer firestoreClient.Close()

	// ── Telephony ──────────────────────────────────────────────────────────
	// Wire the mock provider. Replace with a real Twilio provider here later
	// without changing any service or handler code.
	mockTelephony := telephony.NewMockTelephonyProvider()

	// ── Repositories ───────────────────────────────────────────────────────
	agentRepository := repository.NewAgentRepository(
		firestoreClient,
	)

	callRepository := repository.NewCallRepository(
		firestoreClient,
	)

	// ── Services ───────────────────────────────────────────────────────────
	agentService := service.NewAgentService(
		agentRepository,
	)

	callService := service.NewCallService(
		callRepository,
		agentRepository,
		mockTelephony,
	)

	// ── Handlers ───────────────────────────────────────────────────────────
	agentHandler := handler.NewAgentHandler(
		agentService,
	)

	callHandler := handler.NewCallHandler(
		callService,
	)

	mockTelephonyHandler := handler.NewMockTelephonyHandler(mockTelephony)

	// ── Router ─────────────────────────────────────────────────────────────
	r := gin.Default()

	// Allow the React dev server (and any other origin) to call the API.
	// In production this should be restricted to the actual frontend origin.
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Core DR endpoints.
	r.POST("/dr/enqueue", callHandler.Enqueue)
	r.GET("/dr/queue", callHandler.Queue)
	r.POST("/dr/accept", callHandler.Accept)
	r.POST("/dr/call-status", callHandler.CallStatus)

	r.POST("/dr/agents", agentHandler.CreateAgent)
	r.GET("/dr/agents/:id", agentHandler.GetAgent)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Development-only: inspect mock Twilio state.
	r.GET("/mock/telephony/conferences", mockTelephonyHandler.ListConferences)
	r.GET("/mock/telephony/conferences/:name", mockTelephonyHandler.GetConference)

	log.Println("server running on :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
