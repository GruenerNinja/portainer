package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/endpointutils"
	"github.com/portainer/portainer/api/logs"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/gorilla/websocket"
)

const maxLogTail = 10_000

var logResourceIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

// @summary Stream Docker logs over a websocket
// @description Streams container, service, or task logs through the existing access-controlled Docker proxy.
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags websocket
// @produce octet-stream
// @param endpointId query int true "Environment identifier"
// @param resource query string true "Docker resource type" Enums(containers,services,tasks)
// @param id query string true "Docker resource identifier"
// @param nodeName query string false "Swarm node name"
// @param timestamps query bool false "Include timestamps"
// @param since query int false "Unix timestamp"
// @param tail query int false "Initial line count (maximum 10000)"
// @success 101
// @failure 400
// @failure 403
// @failure 404
// @failure 500
// @router /websocket/logs [get]
func (handler *Handler) websocketLogs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: endpointId", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the environment in the database", err)
	}
	if err != nil {
		return httperror.InternalServerError("Unable to find the environment in the database", err)
	}
	if !endpointutils.IsDockerEndpoint(endpoint) {
		return httperror.BadRequest("Environment does not support Docker logs", nil)
	}

	if err := handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint); err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	proxyRequest, err := createDockerLogsProxyRequest(r)
	if err != nil {
		return httperror.BadRequest("Invalid log stream parameters", err)
	}

	proxy := handler.ProxyManager.GetEndpointProxy(endpoint)
	if proxy == nil {
		proxy, err = handler.ProxyManager.CreateAndRegisterEndpointProxy(endpoint)
		if err != nil {
			return httperror.InternalServerError("Unable to create environment proxy", err)
		}
	}

	r.Header.Del("Origin")
	conn, err := handler.connectionUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return httperror.InternalServerError("Unable to upgrade log stream", err)
	}
	defer logs.CloseAndLogErr(conn)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	proxyRequest = proxyRequest.WithContext(ctx)

	streamWriter := newWebsocketStreamWriter(conn)
	go streamWriter.watchClient(cancel)
	go streamWriter.keepAlive(ctx)

	proxy.ServeHTTP(streamWriter, proxyRequest)
	return nil
}

func createDockerLogsProxyRequest(r *http.Request) (*http.Request, error) {
	resource := r.FormValue("resource")
	switch resource {
	case "containers", "services", "tasks":
	default:
		return nil, fmt.Errorf("unsupported resource type %q", resource)
	}

	resourceID := r.FormValue("id")
	if !logResourceIDPattern.MatchString(resourceID) {
		return nil, fmt.Errorf("invalid resource identifier")
	}

	timestamps, err := optionalBool(r.FormValue("timestamps"), false)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamps value: %w", err)
	}

	since, err := optionalNonNegativeInt(r.FormValue("since"), 0)
	if err != nil {
		return nil, fmt.Errorf("invalid since value: %w", err)
	}

	tail, err := optionalNonNegativeInt(r.FormValue("tail"), 100)
	if err != nil {
		return nil, fmt.Errorf("invalid tail value: %w", err)
	}
	if tail > maxLogTail {
		tail = maxLogTail
	}

	query := url.Values{}
	query.Set("follow", "1")
	query.Set("stdout", "1")
	query.Set("stderr", "1")
	query.Set("timestamps", strconv.FormatBool(timestamps))
	query.Set("since", strconv.FormatInt(since, 10))
	query.Set("tail", strconv.FormatInt(tail, 10))

	proxyRequest := r.Clone(r.Context())
	proxyRequest.Method = http.MethodGet
	proxyRequest.RequestURI = ""
	proxyRequest.URL = &url.URL{
		Path:     fmt.Sprintf("/%s/%s/logs", resource, resourceID),
		RawQuery: query.Encode(),
	}
	// The authenticated request context is retained by Clone. Forward only the
	// headers the Docker proxy needs so browser cookies, WebSocket headers, and
	// compression negotiation never reach the runtime log endpoint.
	proxyRequest.Header = http.Header{"Accept": []string{"application/octet-stream"}}
	if nodeName := r.FormValue("nodeName"); nodeName != "" {
		proxyRequest.Header.Set(portainer.PortainerAgentTargetHeader, nodeName)
	}

	return proxyRequest, nil
}

func optionalBool(value string, fallback bool) (bool, error) {
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseBool(value)
}

func optionalNonNegativeInt(value string, fallback int64) (int64, error) {
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, fmt.Errorf("value must not be negative")
	}
	return parsed, nil
}

type websocketStreamWriter struct {
	conn       *websocket.Conn
	header     http.Header
	mu         sync.Mutex
	statusCode int
}

func newWebsocketStreamWriter(conn *websocket.Conn) *websocketStreamWriter {
	return &websocketStreamWriter{
		conn:       conn,
		header:     http.Header{},
		statusCode: http.StatusOK,
	}
}

func (writer *websocketStreamWriter) Header() http.Header {
	return writer.header
}

func (writer *websocketStreamWriter) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
}

func (writer *websocketStreamWriter) Write(payload []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if err := writer.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return 0, err
	}

	messageType := websocket.BinaryMessage
	if writer.statusCode >= http.StatusBadRequest {
		messageType = websocket.TextMessage
	}
	if err := writer.conn.WriteMessage(messageType, payload); err != nil {
		return 0, err
	}
	return len(payload), nil
}

func (writer *websocketStreamWriter) Flush() {}

func (writer *websocketStreamWriter) watchClient(cancel context.CancelFunc) {
	defer cancel()
	for {
		if _, _, err := writer.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (writer *websocketStreamWriter) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			writer.mu.Lock()
			err := writer.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
			writer.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}
