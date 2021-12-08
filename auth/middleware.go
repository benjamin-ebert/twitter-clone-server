package auth
//
//import (
//	"net/http"
//	"strings"
//	"wtfTwitter/domain"
//)
//
//type UserMw struct {
//	domain.UserService
//}
//
//func (mw *UserMw) Apply(next http.Handler) http.HandlerFunc {
//	return mw.ApplyFn(next.ServeHTTP)
//}
//
//func (mw *UserMw) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		path := r.URL.Path
//		// If the user is requesting a static asset or image
//		// we will not need to lookup the current user so we skip
//		// doing that.
//		if strings.HasPrefix(path, "/assets/") ||
//			strings.HasPrefix(path, "/images/") {
//			next(w, r)
//			return
//		}
//		cookie, err := r.Cookie("remember_token")
//		if err != nil {
//			next(w, r)
//			return
//		}
//		println(mw.UserService)
//		user, err := mw.UserService.FindUserByRemember(cookie.Value)
//		if err != nil {
//			next(w, r)
//			return
//		}
//		ctx := r.Context()
//		ctx = SetUserInContext(ctx, user)
//		r = r.WithContext(ctx)
//		next(w, r)
//	})
//}
//
//// RequireUserMw assumes that User middleware has already been run
//// otherwise it will no work correctly.
//type RequireUserMw struct {
//	UserMw
//}
//
//// Apply assumes that User middleware has already been run
//// otherwise it will no work correctly.
//func (mw *RequireUserMw) Apply(next http.Handler) http.HandlerFunc {
//	return mw.ApplyFn(next.ServeHTTP)
//}
//
//// ApplyFn assumes that User middleware has already been run
//// otherwise it will no work correctly.
//func (mw *RequireUserMw) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		user := GetUserFromContext(r.Context())
//		if user == nil {
//			http.Redirect(w, r, "/home", http.StatusFound)
//			return
//		}
//		next(w, r)
//	})
//}
