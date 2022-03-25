# Twitter Clone Server

This is the server backend of a simplified replication of the Twitter web app.
It does not aim to ship the full functionality of Twitter's backend, but tries to
provide the most important parts of it in order to be usable.
No web-framework was used. Third party packages used are [go-gorm/gorm](https://github.com/go-gorm/gorm), [gorilla/mux](https://github.com/gorilla/mux)
and [gorilla/csrf](https://github.com/gorilla/csrf). It works with a Postgres database.
The client frontend is built with Angular and can be found [here](https://github.com/benjamin-ebert/twitter-clone-client).

The hosted app can be found [here](https://twitter-clone.benjaminebert.net).

As of now it contains the following features:
- traditional authentication system for registration and login with email / password
- oauth authentication with Github
- create and delete tweets, retweets and replies
- upload and attach images to tweets
- create and update a user profile
- upload a profile avatar and header image
- follow and unfollow users
- like and unlike tweets
- view the home feed
- view suggestions for users to follow
- view a user's tweets grouped by four criteria
- search for users by name or handle

## Development Server

### 1. Client App
First clone the [client frontend](https://github.com/benjamin-ebert/twitter-clone-client).
Make sure you have the [Angular CLI](https://angular.io/cli) installed, then `cd` into 
the project root of the client app and run `ng serve`. By default, the client app will
run on [http://localhost:4200/](http://localhost:4200/).

### 2. Database Connection
When compiling, the server will first look for a `.config.json` that contains database 
connection information. If none is found, it will try to establish a default connection 
with the database settings specified in `.config_example.json`. You can either set up 
a local postgres database with those settings, or use any other settings, provide your 
own `.config.json` and specify you settings there.

### 3. Compile and Run
`cd` into the project root of the server and run `go build -gcflags="-N -l" -o YourAppName`
to compile the app. The compiled executable `YourAppName` will appear in the project root
and can be run with `./YourAppName` (Mac and Linux). By default, the server will listen
on [http://localhost:1111/](http://localhost:1111/). Make sure nothing else is running on
that port. To have the server listening on a different port, provide a `.config.json` and
specify the port number there. In that case, you will also have to change the port setting
in the client app's `/proxy.conf.json` and `/environments/environment.ts`, so the client
app can reach your local server.

### 4. OAuth with Github
If you want to use Github-OAuth locally, first create a new oauth app in your Github account.
Copy your app's id and secret and put both into your local `.config.json`. 
At the bottom of the oauth app form is an input called "Authorization Callback URL".
If you did not change your port settings, this will be [http://localhost:1111/api/oauth/github/callback](http://localhost:1111/api/oauth/github/callback).
If you did change your port, be sure to update the url too. Provide the url in your `.config.json`
in the field `redirect_url`. Recompile and run the app. Try logging in with Github.