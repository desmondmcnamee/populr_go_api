package main

import (
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

func (r *router) Put(path string, handler http.Handler) {
	r.PUT(path, wrapHandler(handler))
}

func (r *router) Delete(path string, handler http.Handler) {
	r.DELETE(path, wrapHandler(handler))
}

func (r *router) Options(path string, handler http.Handler) {
	r.OPTIONS(path, wrapHandler(handler))
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}

func NewRouter() *router {
	return &router{httprouter.New()}
}
