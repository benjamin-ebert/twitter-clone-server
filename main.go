package main

import (
	"wtfTwitter/database"
	"wtfTwitter/http"
)

func main() {
	db := database.NewDB("host=localhost user=postgres dbname=wtf_twitter port=5432 sslmode=disable")
	err := database.Open(db)
	if err != nil {
		panic(err)
	}

	userService := database.NewUserService(db.Gorm)
	twitterService := database.NewTweetService(db.Gorm)
	followService := database.NewFollowService(db.Gorm)
	likeService := database.NewLikeService(db.Gorm)

	server := http.NewServer(userService, twitterService, followService, likeService)

	server.Run()
}