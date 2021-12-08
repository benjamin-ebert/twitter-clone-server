package auth
//
//import (
//	"context"
//	"wtfTwitter/domain"
//)
//
//const (
//	userKey privateKey = "user"
//)
//
//type privateKey string
//
//func SetUserInContext(ctx context.Context, user *domain.User) context.Context {
//	return context.WithValue(ctx, userKey, user)
//}
//
//func GetUserFromContext(ctx context.Context) *domain.User {
//	if temp := ctx.Value(userKey); temp != nil {
//		if user, ok := temp.(*domain.User); ok {
//			return user
//		}
//	}
//	return nil
//}
