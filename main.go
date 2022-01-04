package main

import (
	"wtfTwitter/database"
	"wtfTwitter/http"
	"wtfTwitter/storage"
)

func main() {
	config := LoadConfig()
	dbConfig := config.Database

	db := database.NewDB(dbConfig.ConnectionInfo())
	err := database.Open(db)
	if err != nil {
		panic(err)
	}

	userService := database.NewUserService(db.Gorm, config.HMACKey, config.Pepper)
	twitterService := database.NewTweetService(db.Gorm)
	followService := database.NewFollowService(db.Gorm)
	likeService := database.NewLikeService(db.Gorm)
	imageService := storage.NewImageService()

	server := http.NewServer(userService, twitterService, followService, likeService, imageService)

	server.Run(config.Port)
}