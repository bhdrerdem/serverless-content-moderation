package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/apex/gateway"
	types "github.com/bhdrerdem/serverless-content-moderation"
	"github.com/bhdrerdem/serverless-content-moderation/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func main() {

	shared.SnsService.Init()

	addr := ":3000"
	mode := os.Getenv("GIN_MODE")
	router := setupRouter()
	if mode == "release" {
		gateway.ListenAndServe(addr, router)
	} else {
		http.ListenAndServe(addr, router)
	}
}

func setupRouter() *gin.Engine {

	router := gin.New()

	router.POST("/", handler)
	router.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	return router
}

func handler(ctx *gin.Context) {

	content := &types.Content{}

	if err := ctx.BindJSON(&content); err != nil {
		log.Error().Err(err).Msg("failed to bind body json")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if content.Text == "" {
		log.Error().Msg("text is empty")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Field 'text' cannot be empty."})
		return
	}

	content.ID = uuid.NewString()

	contentStr, err := json.Marshal(content)
	if err != nil {
		log.Error().Err(err).Interface("content", content).Msg("failed to marshal content")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process text"})
		return
	}

	err = shared.SnsService.Publish(string(contentStr))
	if err != nil {
		log.Error().Err(err).Str("text", content.Text).Msg("Failed to publish the text")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process text"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Message processing...",
		"id":      content.ID,
	})
}
