package stats

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/containerd/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
)

type ContainerStats struct {
	Running   int `json:"running"`
	Stopped   int `json:"stopped"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Total     int `json:"total"`
}

// ContainerStatsCache contains the inspected status of each container. Keeping
// the per-container values allows callers to apply access-control filtering
// before aggregating the cached statistics.
type ContainerStatsCache map[string]ContainerStats

type DockerClient interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
}

func CalculateContainerStats(ctx context.Context, cli DockerClient, isSwarm bool, containers []container.Summary) (ContainerStats, error) {
	if isSwarm {
		return CalculateContainerStatsForSwarm(containers), nil
	}

	cachedStats, err := InspectContainerStats(ctx, cli, containers)
	return CalculateContainerStatsFromCache(containers, cachedStats), err
}

// InspectContainerStats inspects containers concurrently and retains their
// individual statistics for later aggregation.
func InspectContainerStats(ctx context.Context, cli DockerClient, containers []container.Summary) (ContainerStatsCache, error) {
	cachedStats := make(ContainerStatsCache, len(containers))

	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	var aggErr error

	for i := range containers {
		id := containers[i].ID

		semaphore <- struct{}{}
		wg.Go(func() {
			defer func() { <-semaphore }()

			containerInspection, err := cli.ContainerInspect(ctx, id)
			stat := ContainerStats{}
			if err != nil {
				if errdefs.IsNotFound(err) {
					// An edge case is reported that Docker can list containers with no names,
					// but when inspecting a container with specific ID and it is not found.
					// In this case, we can safely ignore the error.
					// ref@https://linear.app/portainer/issue/BE-12567/500-error-when-loading-docker-dashboard-in-portainer
					return
				}

				mu.Lock()
				aggErr = errors.Join(aggErr, err)
				stat.Total = 1
				cachedStats[id] = stat
				mu.Unlock()
				return
			}
			stat = getContainerStatus(containerInspection.State)
			stat.Total = 1

			mu.Lock()
			cachedStats[id] = stat
			mu.Unlock()
		})
	}

	wg.Wait()

	return cachedStats, aggErr
}

// CalculateContainerStatsFromCache aggregates previously inspected statistics
// for the supplied containers. Containers absent from the cache are ignored,
// matching the existing behaviour for containers removed during inspection.
func CalculateContainerStatsFromCache(containers []container.Summary, cachedStats ContainerStatsCache) ContainerStats {
	result := ContainerStats{}
	for i := range containers {
		stat, ok := cachedStats[containers[i].ID]
		if !ok {
			continue
		}

		result.Running += stat.Running
		result.Stopped += stat.Stopped
		result.Healthy += stat.Healthy
		result.Unhealthy += stat.Unhealthy
		result.Total += stat.Total
	}

	return result
}

func getContainerStatus(state *container.State) ContainerStats {
	stat := ContainerStats{}
	if state == nil {
		return stat
	}

	switch state.Status {
	case container.StateRunning:
		stat.Running++
	case container.StateExited, container.StateDead:
		stat.Stopped++
	}

	if state.Health != nil {
		switch state.Health.Status {
		case container.Healthy:
			stat.Healthy++
		case container.Unhealthy:
			stat.Unhealthy++
		}
	}

	return stat
}

// This is a temporary workaround to calculate container stats for Swarm
// TODO: Remove this once we have a proper way to calculate container stats for Swarm
func CalculateContainerStatsForSwarm(containers []container.Summary) ContainerStats {
	var running, stopped, healthy, unhealthy int
	for _, container := range containers {
		switch container.State {
		case "running":
			running++
		case "exited", "stopped":
			stopped++
		}

		if strings.Contains(container.Status, "(healthy)") {
			healthy++
		} else if strings.Contains(container.Status, "(unhealthy)") {
			unhealthy++
		}
	}

	return ContainerStats{
		Running:   running,
		Stopped:   stopped,
		Healthy:   healthy,
		Unhealthy: unhealthy,
		Total:     len(containers),
	}
}
