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
  - [x] delete 
  - [x] list user tweets
  - [ ] feed of tweets by followed users
- [x] follow / unfollow
  - [ ] declare proper self-ref. m2m with gorm?
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
  - [ ] cascade deletion (tweets)
  - [ ] report upload progress? https://freshman.tech/file-upload-golang/#report-the-upload-progress
  - [ ] obfuscate location
  - [ ] public file server?
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