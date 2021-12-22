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
- [ ] tweets
  - [x] create
  - [ ] update?
  - [x] delete 
  - [ ] list user tweets
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
- [ ] oauth
- [ ] messages
- [ ] seeder
----
- [ ] clean up
- [ ] comment
- [ ] test 
- [ ] deploy
- [ ] push
- [ ] test build
- [ ] readme