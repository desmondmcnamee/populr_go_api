package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
)

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)

				debug.PrintStack()
				WriteError(w, ErrInternalServer)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

func acceptHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.api+json" {
			WriteError(w, ErrNotAcceptable)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func contentTypeHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/vnd.api+json" {
			log.Printf("Error Bad Content Type: %s", r.Header.Get("Content-Type"))
			WriteError(w, ErrUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func userIdHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("x-key")
		if key == "" {
			log.Printf("No x-key passed")
			WriteError(w, ErrNoXKey)
			return
		}

		if next != nil {
			context.Set(r, "x-key", key)
			next.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(fn)
}

func (c *appContext) newTokenHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		passedToken := r.Header.Get("new-token")
		userId := r.Header.Get("x-key")
		if passedToken == "" {
			log.Printf("No token passed")
			WriteError(w, ErrNoToken)
			return
		}

		var user PhoneTokenUser
		err := c.db.Get(&user, "SELECT id, username, phone_number, new_token FROM users WHERE id=$1", userId)
		if err != nil {
			log.Println("Error checking token: ", err)
			WriteError(w, ErrInternalServer)
			return
		}

		if passedToken != user.NewToken {
			log.Println("Error checking token: ", err)
			WriteError(w, ErrBadToken)
			return
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(fn)
}

func bodyHandler(v interface{}) func(http.Handler) http.Handler {
	t := reflect.TypeOf(v)

	m := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			val := reflect.New(t).Interface()
			err := json.NewDecoder(r.Body).Decode(val)

			if err != nil {
				WriteError(w, ErrBadRequest)
				return
			}

			if next != nil {
				context.Set(r, "body", val)
				next.ServeHTTP(w, r)
			}
		}

		return http.HandlerFunc(fn)
	}

	return m
}
