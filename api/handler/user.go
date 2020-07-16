package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arifseft/clean-architecture-sample/api/middleware"
	"github.com/arifseft/clean-architecture-sample/api/views"
	"github.com/arifseft/clean-architecture-sample/pkg/user"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

func register(svc user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			views.Wrap(views.ErrMethodNotAllowed, w)
			return
		}

		var user user.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			views.Wrap(err, w)
			return
		}

		u, err := svc.Register(r.Context(), &user)
		if err != nil {
			views.Wrap(err, w)
			return
		}
		w.WriteHeader(http.StatusCreated)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": u.Email,
			"id":    u.ID,
		})
		tokenString, err := token.SignedString([]byte(viper.GetString("jwt_secret")))
		if err != nil {
			views.Wrap(err, w)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": tokenString,
			"user":  u,
		})
		return
	})
}

func login(svc user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			views.Wrap(views.ErrMethodNotAllowed, w)
			return
		}
		var user user.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			views.Wrap(err, w)
			return
		}

		u, err := svc.Login(r.Context(), user.Email, user.Password)
		if err != nil {
			views.Wrap(err, w)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": u.Email,
			"id":    u.ID,
		})
		tokenString, err := token.SignedString([]byte(viper.GetString("jwt_secret")))
		if err != nil {
			views.Wrap(err, w)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": tokenString,
			"user":  u,
		})
		return
	})
}

func profile(svc user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// @protected
		// @description build profile
		if r.Method == http.MethodPost {
			var user user.User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				views.Wrap(err, w)
				return
			}

			claims, err := middleware.ValidateAndGetClaims(r.Context())
			if err != nil {
				views.Wrap(err, w)
				return
			}
			user.Email = claims["email"].(string)
			u, err := svc.BuildProfile(r.Context(), &user)
			if err != nil {
				views.Wrap(err, w)
				return
			}

			json.NewEncoder(w).Encode(u)
			return
		} else if r.Method == http.MethodGet {

			// @description view profile
			claims, err := middleware.ValidateAndGetClaims(r.Context())
			if err != nil {
				views.Wrap(err, w)
				return
			}
			u, err := svc.GetUserProfile(r.Context(), claims["email"].(string))
			if err != nil {
				views.Wrap(err, w)
				return
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "User profile",
				"data":    u,
			})
			return
		} else {
			views.Wrap(views.ErrMethodNotAllowed, w)
			return
		}
	})
}

func changePassword(svc user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var u user.User
			if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
				views.Wrap(err, w)
				return
			}

			claims, err := middleware.ValidateAndGetClaims(r.Context())
			if err != nil {
				views.Wrap(err, w)
				return
			}
			if err := svc.ChangePassword(r.Context(), claims["email"].(string), u.Password); err != nil {
				views.Wrap(err, w)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "password changed",
			})
			return
			return
		} else {
			views.Wrap(views.ErrMethodNotAllowed, w)
			return
		}
	})
}

// expose handlers
func MakeUserHandler(r *http.ServeMux, svc user.Service) {
	r.Handle("/api/v1/user/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	}))
	r.Handle("/api/v1/user/register", register(svc))
	r.Handle("/api/v1/user/login", login(svc))
	r.Handle("/api/v1/user/profile", middleware.Validate(profile(svc)))
	r.Handle("/api/v1/user/pwd", middleware.Validate(changePassword(svc)))
}
