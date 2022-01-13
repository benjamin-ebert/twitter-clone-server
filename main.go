package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
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

	// Start app services.
	// TODO: Refactor this with functional options.
	userService := auth.NewUserService(db.Gorm, config.HMACKey, config.Pepper)
	tweetService := crud.NewTweetService(db.Gorm)
	followService := crud.NewFollowService(db.Gorm)
	likeService := crud.NewLikeService(db.Gorm)
	imageService := crud.NewImageService()
	oauthService := crud.NewOAuthService(db.Gorm)
	githubOAuth := &oauth2.Config{
		ClientID:     config.Github.ID,
		ClientSecret: config.Github.Secret,
		RedirectURL:  "http://localhost:1111/oauth/github/callback",
		Endpoint: github.Endpoint,
		//Endpoint: oauth2.Endpoint{
		//	AuthURL: config.Github.AuthURL,
		//	TokenURL: config.Github.TokenURL,
		//},
	}

	// Set up a webserver.
	server := http.NewServer(
		userService,
		tweetService,
		followService,
		likeService,
		imageService,
		oauthService,
		*githubOAuth)

	// Serve the app.
	server.Run(config.Port)
}