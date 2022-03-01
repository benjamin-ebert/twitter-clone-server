# Twitter Clone

latest:
- preload user profile relations
- add belongsto tweet relation to like domain
- hard delete likes
- check if tweet exists before deleting it
  (so it only deletes if the user has created it)

todo:

- separate json structs for spa?
- remove password hash / remember hash from user json?
- fix occasional superfluous response panic - always return something at the end of handlers?
- add validation for new crud / related methods?
- make helper tweet queries concurrent? getting images, association counts, auth bools etc.
- 

- [x] auth system
  - [ ] proper user json fields
- [ ] user 
  - [ ] profile data update
  - [x] avatar and header update
- [ ] tweets
  - [x] create
    - [x] images add
  - [ ] show
  - [x] delete 
  - [ ] validate deletion! only owner can
  - [x] list user tweets
  - [ ] feed of tweets by followed users
- [x] follow / unfollow
  - [x] declare proper self-ref. m2m with gorm?
- [x] reply
- [x] retweet
- [x] like 
  - [x] create
  - [x] validate create
  - [x] delete
- [ ] image uploads
  - [x] basic functionality
  - [x] user avatar
  - [x] user header
  - [x] tweet images
  - [x] creation validation 
    - [ ] only creator can
    - [ ] tweet exists
    - [x] max img count
  - [x] avatar / header deletion 
  - [x] cascade deletion (tweets)
  - [ ] delete validation? only creator can
  - [ ] report upload progress? https://freshman.tech/file-upload-golang/#report-the-upload-progress
  - [ ] obfuscate location
  - [ ] public file server?
- [x] ERROR HANDLING
- [x] oauth
- [x] CSRF
- [x] refactor services construction
- [x] .json.conf.example
- [x] todos in the code comments
----
- [ ] seeder ?
- [x] clean up
- [x] comment
- [ ] test 
- [ ] deploy
- [ ] push
- [ ] test build
- [ ] readme

auth better in this project:
- auth/user.go to crud/user.go
- http/auth.go and separate http/oauth.go
- 
future auth system:
- user model has only id, email and timestamps (and name maybe)
- auth model has password, password hash, remember and remember hash
- oauth model has the oauth stuff
- user-auth 1:1
- user-oauth 1:n
- separate http and crud service stuff for each
- user crud is doing email validation
- auth crud is doing all the string and hashing crap, and password and remember validation
- auth http is doing all the middleware and cookie crap
- user and oauth hook into auth, just like now
- separating user crud and auth crud is probably tricky but doable