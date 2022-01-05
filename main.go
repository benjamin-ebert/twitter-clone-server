package main

import (
	"wtfTwitter/auth"
	"wtfTwitter/crud"
	"wtfTwitter/http"
)

// main is the app's entry point.
func main() {
	// Load configuration (from a .config.json file if present, otherwise use default dev setup).
	config := LoadConfig()
	dbConfig := config.Database

	// Open a database connection.
	db := NewDB(dbConfig.ConnectionInfo())
	err := Open(db)
	if err != nil {
		panic(err)
	}

	// Start necessary app services.
	userService := auth.NewUserService(db.Gorm, config.HMACKey, config.Pepper)
	twitterService := crud.NewTweetService(db.Gorm)
	followService := crud.NewFollowService(db.Gorm)
	likeService := crud.NewLikeService(db.Gorm)
	imageService := crud.NewImageService()

	// Set up a webserver.
	server := http.NewServer(userService, twitterService, followService, likeService, imageService)

	// Serve the app.
	server.Run(config.Port)
}