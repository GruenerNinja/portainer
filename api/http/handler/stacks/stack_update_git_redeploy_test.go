package stacks

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/dataservices/source"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
)

	t.Parallel()

	existing := &gittypes.GitAuthentication{
		Username: "existing-user",
		Password: "existing-pass",
	}

	tests := []struct {
		name    string
		auth    *gittypes.GitAuthentication
		payload stackGitRedeployPayload
		want    gittypes.GitAuthentication
	}{
		{
			name:    "no existing auth, flag off, no creds",
			auth:    nil,
			payload: stackGitRedeployPayload{RepositoryAuthentication: false},
			want:    gittypes.GitAuthentication{},
		},
		{
			name:    "no existing auth, flag off, creds provided",
			auth:    nil,
			payload: stackGitRedeployPayload{RepositoryAuthentication: false, RepositoryPassword: "pass"},
			want:    gittypes.GitAuthentication{},
		},
		{
			name:    "no existing auth, flag on, empty password",
			auth:    nil,
			payload: stackGitRedeployPayload{RepositoryAuthentication: true, RepositoryUsername: "user"},
			want:    gittypes.GitAuthentication{},
		},
		{
			name:    "no existing auth, flag on, password set",
			auth:    nil,
			payload: stackGitRedeployPayload{RepositoryAuthentication: true, RepositoryUsername: "user", RepositoryPassword: "pass"},
			want:    gittypes.GitAuthentication{Username: "user", Password: "pass"},
		},
		{
			name:    "no existing auth, flag on, password set but no username",
			auth:    nil,
			payload: stackGitRedeployPayload{RepositoryAuthentication: true, RepositoryPassword: "pass"},
			want:    gittypes.GitAuthentication{Username: "", Password: "pass"},
		},
		{
			name:    "existing auth, flag off",
			auth:    existing,
			payload: stackGitRedeployPayload{RepositoryAuthentication: false},
			want:    *existing,
		},
		{
			name:    "existing auth, flag on, empty password",
			auth:    existing,
			payload: stackGitRedeployPayload{RepositoryAuthentication: true, RepositoryUsername: "new-user"},
			want:    *existing,
		},
		{
			name:    "existing auth, flag on, password set",
			auth:    existing,
			payload: stackGitRedeployPayload{RepositoryAuthentication: true, RepositoryUsername: "new-user", RepositoryPassword: "new-pass"},
			want:    gittypes.GitAuthentication{Username: "new-user", Password: "new-pass"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := &gittypes.RepoConfig{Authentication: tc.auth}
			got := resolveGitAuthFromRedeployPayload(cfg, tc.payload)
			require.Equal(t, tc.want, got)
		})
	}
}

func setupDeployKubernetesStackInlineTest(t *testing.T, deployErr error, initialStatus portainer.StackStatus) (*Handler, *portainer.Stack, *gittypes.RepoConfig, portainer.SourceID, *security.RestrictedRequestContext) {
	t.Helper()

	var manifest = `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`
	var configHash = "testhash"
	tempDir := t.TempDir()
	require.NoError(t, os.WriteFile(filesystem.JoinPaths(tempDir, "manifest.yml"), []byte(manifest), 0o644))

	_, store := datastore.MustNewTestStore(t, true, false)

	adminUserContext := source.InsecureNewAdminContext()

	src := &portainer.Source{
		Type: portainer.SourceTypeGit,
		Git:  &gittypes.GitSource{URL: "https://example.com/repo.git"},
	}
	require.NoError(t, store.Source().Create(adminUserContext, src))

	stack := &portainer.Stack{
		ID:          1,
		Name:        "k8s-stack",
		Type:        portainer.KubernetesStack,
		EndpointID:  1,
		Namespace:   "default",
		ProjectPath: tempDir,
		EntryPoint:  "manifest.yml",
		Status:      initialStatus,
	}

	wf := &portainer.Workflow{
		Artifacts: []portainer.Artifact{{
			StackID: stack.ID,
			Files:   []portainer.ArtifactFile{{SourceID: src.ID, Ref: "refs/heads/main", Hash: configHash}},
		}},
	}
	require.NoError(t, store.Workflow().Create(wf))
	stack.WorkflowID = wf.ID

	require.NoError(t, store.Stack().Create(stack))

	handler := &Handler{
		DataStore:          store,
		KubernetesDeployer: &stubKubernetesDeployer{deployErr: deployErr},
		stackCreationMutex: &sync.Mutex{},
	}


}

	t.Parallel()

	t.Parallel()


}

	t.Parallel()


}
