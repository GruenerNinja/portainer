package stacks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices/source"
	"github.com/portainer/portainer/api/datastore"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/portainer/portainer/api/stacks/stackutils"

	"github.com/google/uuid"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackUpdateGitWebhookUniqueness(t *testing.T) {
	t.Parallel()
	webhook, err := uuid.NewRandom()
	require.NoError(t, err)

	_, store := datastore.MustNewTestStore(t, false, false)

	endpoint := &portainer.Endpoint{
		ID:   123,
		Name: "endpoint1",
		Type: portainer.DockerEnvironment,
	}
	err = store.Endpoint().Create(endpoint)
	require.NoError(t, err)

	const stack1ID = portainer.StackID(456)
	const stack2ID = portainer.StackID(457)

	sharedSrc := &portainer.Source{
		Type: portainer.SourceTypeGit,
		Git:  &gittypes.GitSource{URL: "https://github.com/portainer/portainer.git"},
	}
	err = store.Source().Create(source.InsecureNewAdminContext(), sharedSrc)
	require.NoError(t, err)

	wf1 := &portainer.Workflow{Artifacts: []portainer.Artifact{{
		StackID: stack1ID,
		Files:   []portainer.ArtifactFile{{SourceID: sharedSrc.ID}},
	}}}
	err = store.Workflow().Create(wf1)
	require.NoError(t, err)

	wf2 := &portainer.Workflow{Artifacts: []portainer.Artifact{{
		StackID: stack2ID,
		Files:   []portainer.ArtifactFile{{SourceID: sharedSrc.ID}},
	}}}
	err = store.Workflow().Create(wf2)
	require.NoError(t, err)

	stack1 := portainer.Stack{
		ID:         stack1ID,
		EndpointID: endpoint.ID,
		WorkflowID: wf1.ID,
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook: webhook.String(),
		},
	}

	err = store.Stack().Create(&stack1)
	require.NoError(t, err)

	stack2 := stack1
	stack2.ID = stack2ID
	stack2.AutoUpdate = nil
	stack2.WorkflowID = wf2.ID

	err = store.Stack().Create(&stack2)
	require.NoError(t, err)

	handler := NewHandler(testhelpers.NewTestRequestBouncer(), nil)
	handler.DataStore = store

	payload := &stackGitUpdatePayload{
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook: webhook.String(),
		},
	}

	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	url := "/stacks/" + strconv.Itoa(int(stack2.ID)) + "/git?endpointId=" + strconv.Itoa(int(endpoint.ID))
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(jsonPayload))

	rrc := &security.RestrictedRequestContext{
		IsAdmin: true,
		UserID:  1,
		User:    &portainer.User{ID: 1, Role: portainer.AdministratorRole},
	}
	req = req.WithContext(security.StoreRestrictedRequestContext(req, rrc))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestStackUpdateGitRelativePathSettings(t *testing.T) {
	t.Parallel()

	t.Run("preserves existing setting when payload omits relative path fields", func(t *testing.T) {
		t.Parallel()

		handler, store, stack, endpoint, user := setupStackUpdateGitRelativePathTest(t)
		stack.SupportRelativePath = true
		stack.FilesystemPath = "/mnt/stacks"
		require.NoError(t, store.Stack().Update(stack.ID, stack))

		payload := map[string]any{
			"ConfigFilePath":          "docker-compose.yml",
			"RepositoryReferenceName": "refs/heads/main",
			"env":                     []portainer.Pair{},
		}

		rr := serveStackUpdateGit(t, handler, stack, endpoint, user, payload)

		require.Equal(t, http.StatusOK, rr.Code)
		updated, err := store.Stack().Read(stack.ID)
		require.NoError(t, err)
		assert.True(t, updated.SupportRelativePath)
		assert.Equal(t, "/mnt/stacks", updated.FilesystemPath)
	})

	t.Run("enables relative path settings", func(t *testing.T) {
		t.Parallel()

		handler, store, stack, endpoint, user := setupStackUpdateGitRelativePathTest(t)
		payload := map[string]any{
			"ConfigFilePath":          "docker-compose.yml",
			"RepositoryReferenceName": "refs/heads/main",
			"SupportRelativePath":     true,
			"FilesystemPath":          " /mnt/stacks ",
			"env":                     []portainer.Pair{},
		}

		rr := serveStackUpdateGit(t, handler, stack, endpoint, user, payload)

		require.Equal(t, http.StatusOK, rr.Code)
		updated, err := store.Stack().Read(stack.ID)
		require.NoError(t, err)
		assert.True(t, updated.SupportRelativePath)
		assert.Equal(t, "/mnt/stacks", updated.FilesystemPath)
	})

	t.Run("requires filesystem path when enabling relative paths", func(t *testing.T) {
		t.Parallel()

		handler, _, stack, endpoint, user := setupStackUpdateGitRelativePathTest(t)
		payload := map[string]any{
			"ConfigFilePath":          "docker-compose.yml",
			"RepositoryReferenceName": "refs/heads/main",
			"SupportRelativePath":     true,
			"FilesystemPath":          " ",
			"env":                     []portainer.Pair{},
		}

		rr := serveStackUpdateGit(t, handler, stack, endpoint, user, payload)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func setupStackUpdateGitRelativePathTest(t *testing.T) (*Handler, *datastore.Store, *portainer.Stack, *portainer.Endpoint, *portainer.User) {
	t.Helper()

	_, store := datastore.MustNewTestStore(t, false, false)

	user := &portainer.User{ID: 1, Username: "admin", Role: portainer.AdministratorRole}
	require.NoError(t, store.User().Create(user))

	endpoint := &portainer.Endpoint{
		ID:   123,
		Name: "endpoint1",
		Type: portainer.DockerEnvironment,
	}
	require.NoError(t, store.Endpoint().Create(endpoint))

	stackID := portainer.StackID(456)
	src := &portainer.Source{
		Type: portainer.SourceTypeGit,
		Git:  &gittypes.GitSource{URL: "https://github.com/portainer/portainer.git"},
	}
	require.NoError(t, store.Source().Create(source.InsecureNewAdminContext(), src))

	wf := &portainer.Workflow{Artifacts: []portainer.Artifact{{
		StackID: stackID,
		Files: []portainer.ArtifactFile{{
			SourceID: src.ID,
			Path:     "docker-compose.yml",
			Ref:      "refs/heads/main",
			Hash:     "abc123",
		}},
	}}}
	require.NoError(t, store.Workflow().Create(wf))

	stack := &portainer.Stack{
		ID:         stackID,
		Name:       "test-stack",
		Type:       portainer.DockerComposeStack,
		EndpointID: endpoint.ID,
		WorkflowID: wf.ID,
		EntryPoint: "docker-compose.yml",
	}
	require.NoError(t, store.Stack().Create(stack))

	resourceControl := &portainer.ResourceControl{
		ID:                 portainer.ResourceControlID(stack.ID),
		ResourceID:         stackutils.ResourceControlID(stack.EndpointID, stack.Name),
		Type:               portainer.StackResourceControl,
		AdministratorsOnly: false,
	}
	require.NoError(t, store.ResourceControl().Create(resourceControl))

	handler := NewHandler(testhelpers.NewTestRequestBouncer(), nil)
	handler.DataStore = store

	return handler, store, stack, endpoint, user
}

func serveStackUpdateGit(t *testing.T, handler *Handler, stack *portainer.Stack, endpoint *portainer.Endpoint, user *portainer.User, payload any) *httptest.ResponseRecorder {
	t.Helper()

	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	url := "/stacks/" + strconv.Itoa(int(stack.ID)) + "/git?endpointId=" + strconv.Itoa(int(endpoint.ID))
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(jsonPayload))
	rrc := &security.RestrictedRequestContext{
		IsAdmin: user.Role == portainer.AdministratorRole,
		UserID:  user.ID,
		User:    user,
	}
	req = req.WithContext(security.StoreRestrictedRequestContext(req, rrc))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}
