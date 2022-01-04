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

sucks:
- auth package
- database.user needs stuff from auth package
- http.auth needs stuff from auth package
- server is really just a router with half an auth system
- database is really a crud package

better:
- package main
  - main
  - server
  - postgres
  - config
- package auth
  - hmac
  - token
  - service -> contains auth logic and has access to crud.user
- package database -> package crud
  - gets its hmac/pepper shit directly at construction, provided by main.config
- package server -> package http / router / handler
- ???

might work like this:
- stuff from the auth package into database/user
- stuff from http/auth into http/user
- database/postgres to ./postgres (like config)
- rename database/ to crud/
- http/server to ./server?
- rename http/user to http/auth?
- or put http/user into auth/user?