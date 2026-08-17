package tags

import (
	"net/http"

	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
)

// Handler is the HTTP handler used to handle tag operations.
type Handler struct {
	// Embedding gives Handler the methods of mux.Router, similar to delegation in Java.
	*mux.Router
	DataStore dataservices.DataStore
}

// NewHandler creates a handler to manage tag operations.
func NewHandler(bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}
	// The bouncer wraps each endpoint with its access rule before the route is registered.
	h.Handle("/tags",
		bouncer.AdminAccess(httperror.LoggerHandler(h.tagCreate))).Methods(http.MethodPost)
	h.Handle("/tags",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.tagList))).Methods(http.MethodGet)
	h.Handle("/tags/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.tagDelete))).Methods(http.MethodDelete)

	return h
}
