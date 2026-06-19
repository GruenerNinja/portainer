package stackbuilders

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/gitops/workflows"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubFileService satisfies portainer.FileService for git builder tests.
type stubFileService struct {
	portainer.FileService
	root string
}

func (s *stubFileService) GetStackProjectPath(stackIdentifier string) string {
	if s.root != "" {
		return filepath.Join(s.root, stackIdentifier)
	}

	return "/data/compose/" + stackIdentifier
}

type gitServiceWritingFiles struct {
	portainer.GitService
	files      map[string]string
	commitHash string
}

func (g gitServiceWritingFiles) CloneRepository(_ context.Context, destination, _, _, _, _ string, _ bool) error {
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}

	for name, content := range g.files {
		path := filepath.Join(destination, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (g gitServiceWritingFiles) LatestCommitID(_ context.Context, _, _, _, _ string, _ bool) (string, error) {
	return g.commitHash, nil
}

func newGitMethodBuilder(t *testing.T, commitHash string) *GitMethodStackBuilder {
	t.Helper()
	_, store := datastore.MustNewTestStore(t, false, false)
	require.NoError(t, store.User().Create(&portainer.User{ID: 1, Username: "testuser"}))
	return &GitMethodStackBuilder{
		StackBuilder: StackBuilder{
			stack:       &portainer.Stack{},
			fileService: &stubFileService{},
			dataStore:   store,
		},
		gitService: testhelpers.NewGitService(nil, commitHash),
	}
}

func newGitMethodBuilderWithFiles(t *testing.T, files map[string]string) *GitMethodStackBuilder {
	t.Helper()
	_, store := datastore.MustNewTestStore(t, false, false)
	require.NoError(t, store.User().Create(&portainer.User{ID: 1, Username: "testuser"}))
	return &GitMethodStackBuilder{
		StackBuilder: StackBuilder{
			stack:       &portainer.Stack{},
			fileService: &stubFileService{root: t.TempDir()},
			dataStore:   store,
		},
		gitService: gitServiceWritingFiles{
			files:      files,
			commitHash: "feedcafe",
		},
	}
}

func TestGitMethodStackBuilder_WithSourceID_ReferencesExistingSource(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilder(t, "abc123")
	builder.stack.ID = 1

	src := &portainer.Source{
		Name: "my-repo",
		Type: portainer.SourceTypeGit,
		Git: &gittypes.RepoConfig{
			URL: "https://github.com/org/private-repo",
			Authentication: &gittypes.GitAuthentication{
				Username: "git-user",
				Password: "git-token",
			},
		},
	}
	require.NoError(t, builder.dataStore.Source().Create(src))

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			SourceID:      src.ID,
			ReferenceName: "refs/heads/main",
		},
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))
	require.NoError(t, err)

	// Workflow Artifact must reference the existing Source — not a new one.
	referencedSourceID := builderWorkflowSourceID(t, builder)
	assert.Equal(t, src.ID, referencedSourceID)

	// Only one Source exists — no duplicate was created.
	allSources, err := builder.dataStore.Source().ReadAll()
	require.NoError(t, err)
	assert.Len(t, allSources, 1)

	// The merged git config picks up the Source URL/auth.
	readSrc, artifact, err := workflows.GitSourceAndArtifactForStack(builder.dataStore, builder.stack.WorkflowID, builder.stack.ID)
	require.NoError(t, err)
	merged := workflows.MergeSourceAndFile(readSrc, artifact)
	assert.Equal(t, "https://github.com/org/private-repo", merged.URL)
	assert.Equal(t, "refs/heads/main", merged.ReferenceName)
	require.NotNil(t, merged.Authentication)
	assert.Equal(t, "git-user", merged.Authentication.Username)
}

func TestGitMethodStackBuilder_WithMissingSourceID_ReturnsError(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilder(t, "abc123")
	builder.stack.ID = 2

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			SourceID: portainer.SourceID(999), // does not exist
		},
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))
	require.Error(t, err)
}

func TestGitMethodStackBuilder_WithoutSourceID_InlinePathStillWorks(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilder(t, "feedcafe")
	builder.stack.ID = 4

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			URL:           "https://github.com/org/public-repo",
			ReferenceName: "refs/heads/main",
		},
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))
	require.NoError(t, err)

	// A Source was created via the inline path.
	allSources, err := builder.dataStore.Source().ReadAll()
	require.NoError(t, err)
	assert.Len(t, allSources, 1)
	assert.Equal(t, "https://github.com/org/public-repo", allSources[0].Git.URL)
}

func TestGitMethodStackBuilder_PreparePersistsRelativePathSettings(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilder(t, "feedcafe")
	builder.stack.ID = 5

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			URL:           "https://github.com/org/public-repo",
			ReferenceName: "refs/heads/main",
		},
		SupportRelativePath: true,
		FilesystemPath:      "/mnt/stacks",
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))
	require.NoError(t, err)

	assert.True(t, builder.stack.SupportRelativePath)
	assert.Equal(t, "/mnt/stacks", builder.stack.FilesystemPath)
}

func TestGitMethodStackBuilder_PortainerConfigTargetNameResolvesUnderFilesystemPath(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilderWithFiles(t, map[string]string{
		".portainer.yml": "version: 1\ndeploy:\n  mode: flat\n  targetName: tmc-proxy\n",
	})
	builder.stack.ID = 6

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			URL:           "https://github.com/org/public-repo",
			ReferenceName: "refs/heads/main",
		},
		SupportRelativePath: true,
		FilesystemPath:      "/data/compose",
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))
	require.NoError(t, err)

	assert.True(t, builder.stack.SupportRelativePath)
	assert.Equal(t, "/data/compose/tmc-proxy", builder.stack.FilesystemPath)
}

func TestGitMethodStackBuilder_PortainerConfigComposeFilesOverridePayload(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilderWithFiles(t, map[string]string{
		".portainer.yml": "version: 1\ncompose:\n  files:\n    - compose.yml\n    - compose.prod.yml\n",
	})
	builder.stack.ID = 7
	builder.stack.EntryPoint = "docker-compose.yml"

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			URL:           "https://github.com/org/public-repo",
			ReferenceName: "refs/heads/main",
		},
		ComposeFile:     "docker-compose.yml",
		AdditionalFiles: []string{"docker-compose.override.yml"},
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))
	require.NoError(t, err)

	assert.Equal(t, "compose.yml", builder.stack.EntryPoint)
	assert.Equal(t, []string{"compose.prod.yml"}, builder.stack.AdditionalFiles)

	readSrc, artifact, err := workflows.GitSourceAndArtifactForStack(builder.dataStore, builder.stack.WorkflowID, builder.stack.ID)
	require.NoError(t, err)
	merged := workflows.MergeSourceAndFile(readSrc, artifact)
	assert.Equal(t, "compose.yml", merged.ConfigFilePath)
}

func TestGitMethodStackBuilder_PortainerConfigDeployRequiresRelativePath(t *testing.T) {
	t.Parallel()
	builder := newGitMethodBuilderWithFiles(t, map[string]string{
		".portainer.yml": "version: 1\ndeploy:\n  mode: flat\n  targetName: tmc-proxy\n",
	})
	builder.stack.ID = 8

	payload := &StackPayload{
		RepositoryConfigPayload: RepositoryConfigPayload{
			URL:           "https://github.com/org/public-repo",
			ReferenceName: "refs/heads/main",
		},
	}

	err := builder.prepare(context.Background(), payload, portainer.UserID(1))

	require.Error(t, err)
	assert.Contains(t, err.Error(), ".portainer.yml deploy config requires relative path volumes")
}

// builderWorkflowSourceID returns the first SourceID referenced by the Workflow Artifact for this stack.
func builderWorkflowSourceID(t *testing.T, builder *GitMethodStackBuilder) portainer.SourceID {
	t.Helper()
	require.NotZero(t, builder.stack.WorkflowID)

	wf, err := builder.dataStore.Workflow().Read(builder.stack.WorkflowID)
	require.NoError(t, err)
	require.Len(t, wf.Artifacts, 1)
	require.Len(t, wf.Artifacts[0].Files, 1)
	return wf.Artifacts[0].Files[0].SourceID
}
