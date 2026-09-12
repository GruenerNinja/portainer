package websocket

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/kubernetes/cli"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Handler is the HTTP handler used to handle websocket operations.
type Handler struct {
	*mux.Router
	DataStore                   dataservices.DataStore
	SignatureService            portainer.DigitalSignatureService
	ReverseTunnelService        portainer.ReverseTunnelService
	KubernetesClientFactory     *cli.ClientFactory
	ProxyManager                *proxy.Manager
	requestBouncer              security.BouncerService
	connectionUpgrader          websocket.Upgrader
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
	eventHub                    *eventHub
}

// NewHandler creates a handler to manage websocket operations.
func NewHandler(kubernetesTokenCacheManager *kubernetes.TokenCacheManager, bouncer security.BouncerService) *Handler {
	h := &Handler{
		Router:                      mux.NewRouter(),
		connectionUpgrader:          websocket.Upgrader{},
		requestBouncer:              bouncer,
		kubernetesTokenCacheManager: kubernetesTokenCacheManager,
		eventHub:                    newEventHub(),
	}
	h.PathPrefix("/websocket/exec").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketExec)))
	h.PathPrefix("/websocket/attach").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketAttach)))
	h.PathPrefix("/websocket/pod").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketPodExec)))
	h.PathPrefix("/websocket/kubernetes-shell").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketShellPodExec)))
	h.Handle("/websocket/events",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketEvents))).Methods("GET")
	h.Handle("/websocket/logs",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketLogs))).Methods("GET")
	return h
}

// PublishMutation notifies connected browsers that cached API data may have
// changed. The event deliberately carries no resource names or identifiers.
func (h *Handler) PublishMutation() {
	h.eventHub.publish()
}
