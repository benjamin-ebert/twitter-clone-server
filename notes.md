# Twitter Clone

latest:
- preload user profile relations
- add belongsto tweet relation to like domain
- hard delete likes
- check if tweet exists before deleting it
  (so it only deletes if the user has created it)

todo:

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
- [ ] ERROR HANDLING
- [ ] oauth
- [ ] seeder ?
----
- [ ] clean up
- [ ] comment
- [ ] test 
- [ ] deploy
- [ ] push
- [ ] test build
- [ ] readme

auth:
- so it's pretty okay as it is now
- but maybe even better:
  - folder / package auth, containing:
    - user.go
    - http.go
    - middleware.go
    - hmac.go
    - token.go
  - user needs access to gorm.db, depends on domain/user.go
  - have an authservice start up in main?
  - server needs the authservice?
  - server would then just pull in the auth routes from auth/http
  - and the middleware from auth/middleware