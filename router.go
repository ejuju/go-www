package httputils

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router interface {
	http.Handler
	HandleEndpoint(method string, path string, handlerFunc http.HandlerFunc)
	HandlePrefix(path string, handler http.Handler)
	ParseURIParams(*http.Request) map[string]string
}

type GorillaRouter struct {
	router *mux.Router
}

func NewGorillaRouter() *GorillaRouter {
	router := mux.NewRouter()
	return &GorillaRouter{router: router}
}

func (gorillaRouter *GorillaRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gorillaRouter.router.ServeHTTP(w, r)
}

func (gorillaRouter *GorillaRouter) HandleEndpoint(path, method string, handlerFunc http.HandlerFunc) {
	gorillaRouter.router.HandleFunc(path, handlerFunc).Methods(method)
}

func (gorillaRouter *GorillaRouter) HandlePrefix(path string, handler http.Handler) {
	gorillaRouter.router.PathPrefix(path).Handler(handler)
}

func (gorillaRouter *GorillaRouter) ParseURIParams(r *http.Request) map[string]string {
	return mux.Vars(r)
}
