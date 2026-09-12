package system

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/client"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/pkg/build"
	libclient "github.com/portainer/portainer/pkg/libhttp/client"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
	"github.com/portainer/portainer/pkg/schedule"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/encoding/json"
)

const versionCheckInterval = 6 * time.Hour

var cachedLatestVersion atomic.Pointer[string]

type versionResponse struct {
	// Whether portainer has an update available
	UpdateAvailable bool `json:"UpdateAvailable" example:"false"`
	// The latest version available
	LatestVersion string `json:"LatestVersion" example:"2.0.0"`

	ServerVersion   string
	VersionSupport  string `json:"VersionSupport" example:"STS/LTS"`
	ServerEdition   string `json:"ServerEdition" example:"CE/EE"`
	DatabaseVersion string
	Build           build.BuildInfo
	Dependencies    build.DependenciesInfo
	Runtime         build.RuntimeInfo
}

// @id systemVersion
// @summary Check for portainer updates
// @description Check if portainer has an update available
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags system
// @produce json
// @success 200 {object} versionResponse "Success"
// @router /system/version [get]
func (handler *Handler) version(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	isAdmin, err := security.IsAdmin(r)
	if err != nil {
		return httperror.Forbidden("Permission denied to access Portainer", err)
	}

	result := &versionResponse{
		ServerVersion:   serverVersion(build.ImageTag, portainer.APIVersion),
		VersionSupport:  portainer.APIVersionSupport,
		DatabaseVersion: portainer.APIVersion,
		ServerEdition:   portainer.Edition.GetEditionLabel(),
		Build:           build.GetBuildInfo(),
		Dependencies:    build.GetDependenciesInfo(),
	}

	if isAdmin {
		result.Runtime = build.GetRuntimeInfo()
	}

	latestVersion := GetLatestVersion()
	if HasNewerVersion(result.ServerVersion, latestVersion) {
		result.UpdateAvailable = true
		result.LatestVersion = latestVersion
	}

	return response.JSON(w, &result)
}

func StartVersionCheckService(ctx context.Context, versionCheckURL string) {
	if err := libclient.ExternalRequestDisabled(versionCheckURL); err != nil {
		return
	}

	refresh := func() { refreshLatestVersion(versionCheckURL) }

	go refresh()
	go schedule.RunOnInterval(ctx, versionCheckInterval, refresh, nil)
}

func refreshLatestVersion(versionCheckURL string) {
	if err := libclient.ExternalRequestDisabled(versionCheckURL); err != nil {
		log.Debug().Err(err).Msg("External request disabled: Version check")
		return
	}

	body, err := client.Get(versionCheckURL, 5)
	if err != nil {
		log.Debug().Err(err).Msg("couldn't fetch latest Portainer release version")
		return
	}

	var data struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		log.Debug().Err(err).Msg("couldn't parse latest Portainer version")
		return
	}

	latestVersion := ""
	for _, tag := range data.Results {
		if _, valid := parseNumericVersion(tag.Name); !valid {
			continue
		}

		if latestVersion == "" || HasNewerVersion(latestVersion, tag.Name) {
			latestVersion = tag.Name
		}
	}

	if latestVersion == "" {
		log.Debug().Msg("couldn't find a valid maintained Portainer version")
		return
	}

	cachedLatestVersion.Store(&latestVersion)
}

func GetLatestVersion() string {
	tagName := cachedLatestVersion.Load()
	if tagName == nil {
		return ""
	}

	return *tagName
}

func HasNewerVersion(currentVersion, latestVersion string) bool {
	currentParts, valid := parseNumericVersion(currentVersion)
	if !valid {
		log.Debug().Str("version", currentVersion).Msg("current Portainer version isn't a numeric release version")

		return false
	}

	latestParts, valid := parseNumericVersion(latestVersion)
	if !valid {
		log.Debug().Str("version", latestVersion).Msg("latest Portainer version isn't a numeric release version")

		return false
	}

	partCount := max(len(currentParts), len(latestParts))
	for i := range partCount {
		var currentPart, latestPart uint64
		if i < len(currentParts) {
			currentPart = currentParts[i]
		}
		if i < len(latestParts) {
			latestPart = latestParts[i]
		}

		if currentPart != latestPart {
			return currentPart < latestPart
		}
	}

	return false
}

func serverVersion(imageTag, apiVersion string) string {
	if _, valid := parseNumericVersion(imageTag); valid {
		return strings.TrimPrefix(strings.TrimSpace(imageTag), "v")
	}

	return apiVersion
}

func parseNumericVersion(version string) ([]uint64, bool) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil, false
	}

	parsed := make([]uint64, len(parts))
	for i, part := range parts {
		value, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, false
		}
		parsed[i] = value
	}

	return parsed, true
}
