package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"wtfTwitter/crud"
	"wtfTwitter/http"
)

// main is the app's entry point.
func main() {
	// TODO: Do the flag shit.

	// Load configuration (from a .config.json file if present, otherwise use default dev setup).
	config := LoadConfig()

	// Open a database connection.
	dbConfig := config.Database
	db := NewDB(dbConfig.ConnectionInfo())
	err := Open(db, config.IsProd())
	must(err)
	defer Close(db)
	err = AutoMigrate(db)
	must(err)

	// Start app services.
	services, err := crud.NewServices(
		db.Gorm,
		crud.WithUser(config.Pepper, config.HMACKey),
		crud.WithOAuth(),
		crud.WithTweet(),
		crud.WithFollow(),
		crud.WithLike(),
		crud.WithImage(),
	)
	must(err)

	// Create an oauth config for Github.
	githubOAuth := &oauth2.Config{
		ClientID:     config.Github.ID,
		ClientSecret: config.Github.Secret,
		RedirectURL:  config.Github.RedirectURL,
		Endpoint: github.Endpoint,
	}

	// Set up a webserver.
	server := http.NewServer(config.IsProd(), githubOAuth, services)

	// Serve the app.
	server.Run(config.Port)
}

// must is a little helper for shortening the panic instruction.
func must(err error) {
	if err != nil {
		panic(err)
	}
}
