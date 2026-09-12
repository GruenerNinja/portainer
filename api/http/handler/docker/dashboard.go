package docker

import (
	"errors"
	"net/http"

	"github.com/docker/docker/api/types/volume"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/docker/stats"
	"github.com/portainer/portainer/api/http/handler/docker/utils"
	"github.com/portainer/portainer/api/http/middlewares"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/uac"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type imagesCounters struct {
	Total int   `json:"total"`
	Size  int64 `json:"size"`
}

type dashboardResponse struct {
	Containers stats.ContainerStats `json:"containers"`
	Services   int                  `json:"services"`
	Images     imagesCounters       `json:"images"`
	Volumes    int                  `json:"volumes"`
	Networks   int                  `json:"networks"`
	Stacks     int                  `json:"stacks"`
}

// @id dockerDashboard
// @summary Get counters for the dashboard
// @description **Access policy**: restricted
// @tags docker
// @security jwt
// @param environmentId path int true "Environment identifier"
// @accept json
// @produce json
// @success 200 {object} dashboardResponse "Success"
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @router /docker/{environmentId}/dashboard [get]
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	requestContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user details from request context", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environment", err)
	}

	cacheKey := dashboardCacheKey{
		endpointID:  endpoint.ID,
		endpointURL: endpoint.URL,
		agentTarget: r.Header.Get(portainer.PortainerAgentTargetHeader),
	}
	cachedData, err := h.dashboardCache.get(r.Context(), cacheKey)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Docker dashboard data", err)
	}

	var resp dashboardResponse
	err = h.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		user, err := tx.User().Read(requestContext.UserID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve user", err)
		}

		containers, err := uac.FilterByResourceControl(cachedData.containers, user, requestContext.UserMemberships, uac.ContainerResourceControlGetter(tx, endpoint.ID))
		if err != nil {
			return err
		}

		services := cachedData.services
		if cachedData.isSwarmManager {
			services, err = uac.FilterByResourceControl(services, user, requestContext.UserMemberships, uac.ServiceResourceControlGetter(tx, endpoint.ID))
			if err != nil {
				return err
			}
		}

		volumes, err := uac.FilterByResourceControl(cachedData.volumes, user, requestContext.UserMemberships, func(item *volume.Volume) (*portainer.ResourceControl, error) {
			if item == nil {
				return nil, errors.New("Found nil volume in volumes list")
			}
			return uac.VolumeResourceControlGetter(tx, endpoint.ID)(*item)
		})
		if err != nil {
			return err
		}

		networks, err := uac.FilterByResourceControl(cachedData.networks, user, requestContext.UserMemberships, uac.NetworkResourceControlGetter(tx, endpoint.ID))
		if err != nil {
			return err
		}

		stackCount := 0
		if endpoint.SecuritySettings.AllowStackManagementForRegularUsers || requestContext.IsAdmin {
			stacks, err := utils.GetDockerStacks(tx, requestContext, endpoint.ID, containers, services)
			if err != nil {
				return httperror.InternalServerError("Unable to retrieve stacks", err)
			}

			stackCount = len(stacks)
		}

		containersStats := stats.CalculateContainerStatsFromCache(containers, cachedData.containerStats)
		if cachedData.isSwarm {
			containersStats = stats.CalculateContainerStatsForSwarm(containers)
		}

		resp = dashboardResponse{
			Images:     cachedData.images,
			Services:   len(services),
			Containers: containersStats,
			Networks:   len(networks),
			Volumes:    len(volumes),
			Stacks:     stackCount,
		}

		return nil
	})

	return response.TxResponse(w, resp, err)
}
