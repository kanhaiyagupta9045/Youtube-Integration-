package service

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func InitService() (*youtube.Service, error) {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY not found in environment variables")
	}

	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	return youtubeService, nil

}
