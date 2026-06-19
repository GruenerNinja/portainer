package stackbuilders

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/gitops/workflows"
	"github.com/portainer/portainer/api/scheduler"
	"github.com/portainer/portainer/api/stacks/deployments"
	"github.com/portainer/portainer/api/stacks/stackutils"
)

type GitMethodStackBuilder struct {
	StackBuilder
	gitService portainer.GitService
	scheduler  *scheduler.Scheduler
}

func (b *GitMethodStackBuilder) prepare(ctx context.Context, payload *StackPayload, userID portainer.UserID) error {
	b.stack.AdditionalFiles = payload.AdditionalFiles
	b.stack.AutoUpdate = payload.AutoUpdate
	b.stack.SupportRelativePath = payload.SupportRelativePath
	b.stack.FilesystemPath = payload.FilesystemPath

	if err := b.initCreatedBy(userID); err != nil {
		return err
	}

	var repoConfig gittypes.RepoConfig
	var sourceID portainer.SourceID

	if payload.SourceID != 0 {
		src, err := b.dataStore.Source().Read(payload.SourceID)
		if err != nil {
			return fmt.Errorf("failed to read source: %w", err)
		}
		if src.Git == nil {
			return fmt.Errorf("source %d has no git configuration", payload.SourceID)
		}

		repoConfig.URL = src.Git.URL
		repoConfig.Authentication = src.Git.Authentication
		repoConfig.TLSSkipVerify = src.Git.TLSSkipVerify
		repoConfig.ReferenceName = payload.ReferenceName
		sourceID = src.ID
	} else {
		if payload.Authentication {
			repoConfig.Authentication = &gittypes.GitAuthentication{
				Username: payload.Username,
				Password: payload.Password,
			}
		}

		repoConfig.URL = payload.URL
		repoConfig.ReferenceName = payload.ReferenceName
		repoConfig.TLSSkipVerify = payload.TLSSkipVerify
	}

	repoConfig.ConfigFilePath = payload.ComposeFile
	if payload.ComposeFile == "" {
		repoConfig.ConfigFilePath = filesystem.ComposeFileDefaultName
	}

	// If a manifest file is specified (for kube git apps), then use it instead of the default compose file name
	if payload.ManifestFile != "" {
		repoConfig.ConfigFilePath = payload.ManifestFile
	}

	stackFolder := strconv.Itoa(int(b.stack.ID))
	// Set the project path on the disk
	b.stack.ProjectPath = b.fileService.GetStackProjectPath(stackFolder)

	getProjectPath := func() string {
		stackFolder := fmt.Sprintf("%d", b.stack.ID)
		return b.fileService.GetStackProjectPath(stackFolder)
	}

	commitHash, err := stackutils.DownloadGitRepository(ctx, repoConfig, b.gitService, getProjectPath)
	if err != nil {
		return fmt.Errorf("failed to download git repository: %w", err)
	}

	if err := b.applyPortainerStackConfig(&repoConfig); err != nil {
		return err
	}

	// Update the latest commit id
	repoConfig.ConfigHash = commitHash

	var workflowID portainer.WorkflowID

	if err := b.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		file := portainer.ArtifactFile{
			Path: repoConfig.ConfigFilePath,
			Ref:  repoConfig.ReferenceName,
			Hash: repoConfig.ConfigHash,
		}

		if sourceID != 0 {
			file.SourceID = sourceID
		} else {
			repoConfig.URL = gittypes.SanitizeURL(repoConfig.URL)

			src, err := workflows.FindOrCreateGitSource(tx, &portainer.Source{
				Name: gittypes.RepoName(repoConfig.URL),
				Type: portainer.SourceTypeGit,
				Git:  &repoConfig,
			})
			if err != nil {
				return fmt.Errorf("failed to find or create source: %w", err)
			}

			file.SourceID = src.ID
		}

		wf := &portainer.Workflow{
			Name: b.stack.Name,
			Artifacts: []portainer.Artifact{{
				StackID: b.stack.ID,
				Files:   []portainer.ArtifactFile{file},
			}},
		}
		if err := tx.Workflow().Create(wf); err != nil {
			return fmt.Errorf("failed to create workflow: %w", err)
		}

		workflowID = wf.ID

		return nil
	}); err != nil {
		return err
	}

	b.stack.WorkflowID = workflowID

	return nil
}

func (b *GitMethodStackBuilder) applyPortainerStackConfig(repoConfig *gittypes.RepoConfig) error {
	config, found, err := stackutils.LoadPortainerStackConfigForFile(b.stack.ProjectPath, repoConfig.ConfigFilePath)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	if config.HasDeployConfig() {
		if !b.stack.SupportRelativePath || strings.TrimSpace(b.stack.FilesystemPath) == "" {
			return fmt.Errorf("%s deploy config requires relative path volumes and a filesystem path", stackutils.PortainerStackConfigFile)
		}

		if config.Deploy.TargetName != "" {
			b.stack.FilesystemPath = portainerTargetFilesystemPath(b.stack.FilesystemPath, config.Deploy.TargetName)
		}
	}

	if len(config.Compose.Files) > 0 {
		repoConfig.ConfigFilePath = config.Compose.Files[0]
		b.stack.EntryPoint = config.Compose.Files[0]
		b.stack.AdditionalFiles = append([]string{}, config.Compose.Files[1:]...)
	}

	return nil
}

func portainerTargetFilesystemPath(filesystemPath string, targetName string) string {
	destination := strings.TrimSpace(filesystemPath)
	if strings.TrimSpace(targetName) == "" {
		return destination
	}

	destination = filepath.Clean(destination)
	if filepath.Base(destination) == targetName {
		return destination
	}

	return filesystem.JoinPaths(destination, targetName)
}

// postDeploy enables the auto-update scheduler job for the stack if configured,
// and persists the resulting job ID back to the database.
func (b *GitMethodStackBuilder) postDeploy(ctx context.Context, stack *portainer.Stack) error {
	if stack.AutoUpdate == nil || stack.AutoUpdate.Interval == "" {
		return nil
	}

	jobID, err := deployments.StartAutoupdate(ctx, stack.ID,
		stack.AutoUpdate.Interval,
		b.scheduler,
		b.stackDeployer,
		b.dataStore,
		b.gitService)
	if err != nil {
		return err
	}

	return b.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		s, err := tx.Stack().Read(stack.ID)
		if err != nil {
			return fmt.Errorf("Unable to retrieve the stack from the database: %w", err)
		}

		s.AutoUpdate.JobID = jobID

		if err := tx.Stack().Update(s.ID, s); err != nil {
			return fmt.Errorf("Unable to update the stack inside the database: %w", err)
		}

		return nil
	})
}
