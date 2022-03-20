package main

import (
	"flag"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"wtfTwitter/crud"
	"wtfTwitter/http"
)

// main is the app's entry point.
func main() {
	// Check if the flag "-prod" to has been provided. It means that we're running in production.
	productionBool := flag.Bool("prod", false, "Provide this flag in production to ensure that a .config.json file is provided before the application starts.")
	flag.Parse()

	// Load configuration from a .config.json file if present, otherwise use the default dev setup.
	// If *productionBool evaluates to true, that means we're in production. In that case the
	// .config.json file is required and the app will panic if no file is found.
	config := LoadConfig(*productionBool)

	// Open a database connection and execute migrations.
	dbConfig := config.Database
	db := NewDB(dbConfig.ConnectionInfo())
	err := Open(db, config.IsProd())
	must(err)
	defer Close(db)
	err = AutoMigrate(db)
	must(err)

	// Start the crud services.
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

	// Create an oauth config object for doing oauth with Github.
	githubOAuth := &oauth2.Config{
		ClientID:     config.Github.ID,
		ClientSecret: config.Github.Secret,
		RedirectURL:  config.Github.RedirectURL,
		Endpoint:     github.Endpoint,
	}

	// Set up a webserver.
	server := http.NewServer(config.IsProd(), config.ClientUrl, githubOAuth, services)

	// Serve the app.
	server.Run(config.Port)
}

// must is a little helper for shortening the panic instruction.
func must(err error) {
	if err != nil {
		panic(err)
	}
}
