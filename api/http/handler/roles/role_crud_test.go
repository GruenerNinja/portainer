package roles

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"

	"github.com/gorilla/mux"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/require"
)

func TestRoleCRUD(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)
	for id := portainer.RoleIDEndpointAdmin; id <= portainer.RoleIDOperator; id++ {
		require.NoError(t, store.Role().Update(portainer.RoleID(id), &portainer.Role{ID: portainer.RoleID(id), Name: "built-in"}))
	}
	handler := &Handler{DataStore: store}

	createBody := bytes.NewBufferString(`{"Name":"Automation reader","Description":"Agent account","Priority":5,"Authorizations":{"DockerStackList":true,"DockerStackCreate":false}}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/roles", createBody)
	createResponse := httptest.NewRecorder()
	require.Nil(t, handler.roleCreate(createResponse, createRequest))

	var created portainer.Role
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &created))
	require.Equal(t, portainer.RoleID(6), created.ID)
	require.True(t, created.Authorizations[portainer.Authorization("DockerStackList")])
	_, containsFalse := created.Authorizations[portainer.Authorization("DockerStackCreate")]
	require.False(t, containsFalse)

	updateBody := bytes.NewBufferString(`{"Name":"Automation stack reader","Description":"Updated","Priority":6,"Authorizations":{"DockerStackList":true}}`)
	updateRequest := mux.SetURLVars(httptest.NewRequest(http.MethodPut, "/roles/6", updateBody), map[string]string{"id": "6"})
	updateResponse := httptest.NewRecorder()
	require.Nil(t, handler.roleUpdate(updateResponse, updateRequest))
	updated, err := store.Role().Read(created.ID)
	require.NoError(t, err)
	require.Equal(t, "Automation stack reader", updated.Name)

	deleteRequest := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/roles/6", nil), map[string]string{"id": "6"})
	require.Nil(t, handler.roleDelete(httptest.NewRecorder(), deleteRequest))
	_, err = store.Role().Read(created.ID)
	require.True(t, store.IsErrObjectNotFound(err))
}

func TestRoleDeleteProtectsBuiltInAndAssignedRoles(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)
	handler := &Handler{DataStore: store}

	builtInRequest := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/roles/1", nil), map[string]string{"id": "1"})
	handlerError := handler.roleDelete(httptest.NewRecorder(), builtInRequest)
	require.NotNil(t, handlerError)
	require.Equal(t, http.StatusForbidden, handlerError.StatusCode)

	role := &portainer.Role{ID: 6, Name: "Assigned", Priority: 5, Authorizations: portainer.Authorizations{}}
	require.NoError(t, store.Role().Update(role.ID, role))
	require.NoError(t, store.Endpoint().Create(&portainer.Endpoint{
		ID: 1, Name: "local", UserAccessPolicies: portainer.UserAccessPolicies{1: {RoleID: role.ID}},
	}))

	assignedRequest := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/roles/6", nil), map[string]string{"id": "6"})
	handlerError = handler.roleDelete(httptest.NewRecorder(), assignedRequest)
	require.NotNil(t, handlerError)
	require.Equal(t, http.StatusConflict, handlerError.StatusCode)
}
