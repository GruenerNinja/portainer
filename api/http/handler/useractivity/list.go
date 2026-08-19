package useractivity

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type listResponse struct {
	Logs       []portainer.ActivityLog `json:"logs"`
	TotalCount int                     `json:"totalCount"`
}

func (handler *Handler) filtered(r *http.Request) ([]portainer.ActivityLog, int, error) {
	q := r.URL.Query()
	keyword := strings.ToLower(q.Get("keyword"))
	after, _ := strconv.ParseInt(q.Get("after"), 10, 64)
	before, _ := strconv.ParseInt(q.Get("before"), 10, 64)
	logs, err := handler.DataStore.ActivityLog().ReadAll(func(entry portainer.ActivityLog) bool {
		if after != 0 && entry.Timestamp < after || before != 0 && entry.Timestamp > before {
			return false
		}
		return keyword == "" || strings.Contains(strings.ToLower(entry.Action), keyword) ||
			strings.Contains(strings.ToLower(entry.Context), keyword) || strings.Contains(strings.ToLower(entry.Username), keyword)
	})
	if err != nil {
		return nil, 0, err
	}

	sortBy, desc := q.Get("sortBy"), q.Get("sortDesc") == "true"
	if sortBy == "" {
		sortBy, desc = "Timestamp", true
	}
	sort.SliceStable(logs, func(i, j int) bool {
		comparison := 0
		switch sortBy {
		case "Action":
			comparison = strings.Compare(logs[i].Action, logs[j].Action)
		case "Context":
			comparison = strings.Compare(logs[i].Context, logs[j].Context)
		case "Username":
			comparison = strings.Compare(logs[i].Username, logs[j].Username)
		default:
			if logs[i].Timestamp < logs[j].Timestamp {
				comparison = -1
			} else if logs[i].Timestamp > logs[j].Timestamp {
				comparison = 1
			}
		}
		if comparison == 0 {
			comparison = int(logs[i].ID - logs[j].ID)
		}
		if desc {
			return comparison > 0
		}
		return comparison < 0
	})

	total := len(logs)
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if offset > len(logs) {
		offset = len(logs)
	}
	logs = logs[offset:]
	if limit > 0 && limit < len(logs) {
		logs = logs[:limit]
	}
	return logs, total, nil
}

// @id UserActivityLogs
// @summary List user activity logs
// @description Lists retained authenticated state-changing API requests. Request bodies and credentials are never recorded.
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} listResponse "Success"
// @router /useractivity/logs [get]
func (handler *Handler) list(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	logs, total, err := handler.filtered(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user activity logs", err)
	}
	return response.JSON(w, listResponse{Logs: logs, TotalCount: total})
}

func (handler *Handler) exportCSV(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	logs, _, err := handler.filtered(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user activity logs", err)
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="portainer-activity-logs.csv"`)
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"Timestamp", "Username", "Action", "Context", "Payload"}); err != nil {
		return httperror.InternalServerError("Unable to export user activity logs", err)
	}
	for _, entry := range logs {
		if err := writer.Write([]string{strconv.FormatInt(entry.Timestamp, 10), entry.Username, entry.Action, entry.Context, entry.Payload}); err != nil {
			return httperror.InternalServerError("Unable to export user activity logs", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return httperror.InternalServerError(fmt.Sprintf("Unable to export %d user activity logs", len(logs)), err)
	}
	return nil
}
