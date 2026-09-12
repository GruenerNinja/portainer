package registries

import (
	"context"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
	"github.com/portainer/portainer/pkg/liboras"
)

type registryCatalogResponse struct {
	Repositories []string `json:"repositories"`
}

type registryTagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// @id RegistryRepositoriesList
// @summary List registry repositories
// @description List repositories visible with the configured registry credentials.
// @description **Access policy**: restricted
// @tags registries
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Registry identifier"
// @success 200 {object} registryCatalogResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access registry"
// @failure 404 "Registry not found"
// @failure 500 "Server error"
// @router /registries/{id}/v2/_catalog [get]
func (handler *Handler) registryRepositoriesList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registry, handlerErr := handler.registryForBrowse(r)
	if handlerErr != nil {
		return handlerErr
	}

	repositories, err := handler.listRepositories(r.Context(), registry)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve registry repositories", err)
	}

	if repositories == nil {
		repositories = []string{}
	}

	return response.JSON(w, registryCatalogResponse{Repositories: repositories})
}

// @id RegistryRepositoryTagsList
// @summary List repository tags
// @description List tags for a repository using the configured registry credentials.
// @description **Access policy**: restricted
// @tags registries
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Registry identifier"
// @param repository path string true "Repository name"
// @success 200 {object} registryTagsResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access registry"
// @failure 404 "Registry not found"
// @failure 500 "Server error"
// @router /registries/{id}/v2/{repository}/tags/list [get]
func (handler *Handler) registryRepositoryTagsList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	registry, handlerErr := handler.registryForBrowse(r)
	if handlerErr != nil {
		return handlerErr
	}

	repository, err := request.RetrieveRouteVariableValue(r, "repository")
	if err != nil {
		return httperror.BadRequest("Invalid repository route variable", err)
	}

	tags, err := handler.listTags(r.Context(), registry, repository)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve repository tags", err)
	}

	if tags == nil {
		tags = []string{}
	}

	return response.JSON(w, registryTagsResponse{Name: repository, Tags: tags})
}

func (handler *Handler) registryForBrowse(r *http.Request) (*portainer.Registry, *httperror.HandlerError) {
	registryID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return nil, httperror.BadRequest("Invalid registry identifier route variable", err)
	}

	registry, err := handler.DataStore.Registry().Read(portainer.RegistryID(registryID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find a registry with the specified identifier inside the database", err)
	}
	if err != nil {
		return nil, httperror.InternalServerError("Unable to find a registry with the specified identifier inside the database", err)
	}

	return registry, nil
}

func listRegistryRepositories(ctx context.Context, registry *portainer.Registry) ([]string, error) {
	client, err := liboras.CreateClient(*registry)
	if err != nil {
		return nil, err
	}

	return liboras.ListRepositories(ctx, registry, client)
}

func listRegistryTags(ctx context.Context, registry *portainer.Registry, repositoryName string) ([]string, error) {
	client, err := liboras.CreateClient(*registry)
	if err != nil {
		return nil, err
	}

	repository, err := client.Repository(ctx, repositoryName)
	if err != nil {
		return nil, err
	}

	tags := []string{}
	err = repository.Tags(ctx, "", func(page []string) error {
		tags = append(tags, page...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tags, nil
}
