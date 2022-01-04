package main

import (
	"wtfTwitter/auth"
	"wtfTwitter/crud"
	"wtfTwitter/http"
)

func main() {
	config := LoadConfig()
	dbConfig := config.Database

	db := NewDB(dbConfig.ConnectionInfo())
	err := Open(db)
	if err != nil {
		panic(err)
	}

	userService := auth.NewUserService(db.Gorm, config.HMACKey, config.Pepper)

	twitterService := crud.NewTweetService(db.Gorm)
	followService := crud.NewFollowService(db.Gorm)
	likeService := crud.NewLikeService(db.Gorm)
	imageService := crud.NewImageService()

	server := http.NewServer(userService, twitterService, followService, likeService, imageService)

	server.Run(config.Port)
}