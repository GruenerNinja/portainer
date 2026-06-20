package deployments

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	dockerclient "github.com/portainer/portainer/api/docker/client"
	k "github.com/portainer/portainer/api/kubernetes"
)

type BaseStackDeployer interface {
	DeploySwarmStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune, pullImage bool) error
	DeployComposeStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune, forcePullImage, forceRecreate bool) error
	UndeployComposeStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint) error
	DeployKubernetesStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint, user *portainer.User) error
}

type StackDeployer interface {
	BaseStackDeployer
	RemoteStackDeployer
}

type stackDeployer struct {
	lock                *sync.Mutex
	swarmStackManager   portainer.SwarmStackManager
	composeStackManager portainer.ComposeStackManager
	kubernetesDeployer  portainer.KubernetesDeployer
	ClientFactory       *dockerclient.ClientFactory
	dataStore           dataservices.DataStore
}

// NewStackDeployer inits a stackDeployer struct with a SwarmStackManager, a ComposeStackManager and a KubernetesDeployer
func NewStackDeployer(swarmStackManager portainer.SwarmStackManager, composeStackManager portainer.ComposeStackManager,
	kubernetesDeployer portainer.KubernetesDeployer, clientFactory *dockerclient.ClientFactory, dataStore dataservices.DataStore,
) *stackDeployer {
	return &stackDeployer{
		lock:                &sync.Mutex{},
		swarmStackManager:   swarmStackManager,
		composeStackManager: composeStackManager,
		kubernetesDeployer:  kubernetesDeployer,
		ClientFactory:       clientFactory,
		dataStore:           dataStore,
	}
}

func (d *stackDeployer) DeploySwarmStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune, pullImage bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	resolvedStack, err := stackWithResolvedSecrets(ctx, d.dataStore, stack)
	if err != nil {
		return err
	}

	return d.swarmStackManager.Deploy(ctx, resolvedStack, prune, pullImage, endpoint, registries)
}

func (d *stackDeployer) DeployComposeStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune, forcePullImage, forceRecreate bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	resolvedStack, err := stackWithResolvedSecrets(ctx, d.dataStore, stack)
	if err != nil {
		return err
	}

	options := portainer.ComposeOptions{Registries: registries}

	// --force-recreate doesn't pull updated images
	if forcePullImage {
		if err := d.composeStackManager.Pull(ctx, resolvedStack, endpoint, options); err != nil {
			return err
		}
	}

	return d.composeStackManager.Up(ctx, resolvedStack, endpoint, portainer.ComposeUpOptions{
		ComposeOptions: options,
		ForceRecreate:  forceRecreate,
		Prune:          prune,
	})
}

func (d *stackDeployer) UndeployComposeStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.composeStackManager.Down(ctx, stack, endpoint)
}

func (d *stackDeployer) DeployKubernetesStack(ctx context.Context, stack *portainer.Stack, endpoint *portainer.Endpoint, user *portainer.User) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	resolvedStack, err := stackWithResolvedSecrets(ctx, d.dataStore, stack)
	if err != nil {
		return err
	}

	appLabels := k.KubeAppLabels{
		StackID:   int(resolvedStack.ID),
		StackName: resolvedStack.Name,
		Owner:     user.Username,
	}

	if resolvedStack.WorkflowID == 0 {
		appLabels.Kind = "content"
	} else {
		appLabels.Kind = "git"
	}

	k8sDeploymentConfig := CreateKubernetesStackDeploymentConfig(resolvedStack, d.kubernetesDeployer, appLabels, user, endpoint)

	if err := k8sDeploymentConfig.Deploy(ctx); err != nil {
		return errors.Wrap(err, "failed to deploy kubernetes application")
	}

	return nil
}
