package resourcecontrols

import (
	"errors"
	"sort"

	portainer "github.com/portainer/portainer/api"
)

func validateDistinctResourceAccesses(readWriteUsers, readOnlyUsers, readWriteTeams, readOnlyTeams []int) error {
	if hasOverlap(readWriteUsers, readOnlyUsers) {
		return errors.New("invalid payload: a user cannot have both read-write and read-only access")
	}
	if hasOverlap(readWriteTeams, readOnlyTeams) {
		return errors.New("invalid payload: a team cannot have both read-write and read-only access")
	}
	return nil
}

func hasOverlap(left, right []int) bool {
	seen := make(map[int]struct{}, len(left))
	for _, id := range left {
		seen[id] = struct{}{}
	}
	for _, id := range right {
		if _, ok := seen[id]; ok {
			return true
		}
	}
	return false
}

func equalIntSets(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	left = append([]int(nil), left...)
	right = append([]int(nil), right...)
	sort.Ints(left)
	sort.Ints(right)
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func readOnlyAccessIDs(resourceControl *portainer.ResourceControl) ([]int, []int) {
	users := make([]int, 0)
	teams := make([]int, 0)
	for _, access := range resourceControl.UserAccesses {
		if access.AccessLevel == portainer.ReadOnlyAccessLevel {
			users = append(users, int(access.UserID))
		}
	}
	for _, access := range resourceControl.TeamAccesses {
		if access.AccessLevel == portainer.ReadOnlyAccessLevel {
			teams = append(teams, int(access.TeamID))
		}
	}
	sort.Ints(users)
	sort.Ints(teams)
	return users, teams
}
