package useractivity

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
)

type Handler struct {
	*mux.Router
	DataStore dataservices.DataStore
}

func NewHandler(bouncer security.BouncerService, dataStore dataservices.DataStore) *Handler {
	h := &Handler{Router: mux.NewRouter(), DataStore: dataStore}
	h.Handle("/useractivity/logs", bouncer.AdminAccess(httperror.LoggerHandler(h.list))).Methods(http.MethodGet)
	h.Handle("/useractivity/logs.csv", bouncer.AdminAccess(httperror.LoggerHandler(h.exportCSV))).Methods(http.MethodGet)
	return h
}
