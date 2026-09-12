package auth

import (
	"fmt"
	"regexp"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

func (handler *Handler) syncOAuthTeamMemberships(userID portainer.UserID, settings *portainer.OAuthSettings, claims map[string]any) error {
	desiredTeamIDs, managedTeamIDs, err := resolveOAuthTeamMemberships(settings, claims)
	if err != nil {
		return err
	}

	return handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		memberships, err := tx.TeamMembership().TeamMembershipsByUserID(userID)
		if err != nil {
			return fmt.Errorf("load existing memberships: %w", err)
		}

		existingTeamIDs := make(map[portainer.TeamID]struct{}, len(memberships))
		for i := range memberships {
			membership := &memberships[i]
			existingTeamIDs[membership.TeamID] = struct{}{}

			if _, managed := managedTeamIDs[membership.TeamID]; !managed {
				continue
			}
			if _, desired := desiredTeamIDs[membership.TeamID]; desired {
				continue
			}

			if err := tx.TeamMembership().Delete(membership.ID); err != nil {
				return fmt.Errorf("remove membership for team %d: %w", membership.TeamID, err)
			}
		}

		for teamID := range desiredTeamIDs {
			if _, exists := existingTeamIDs[teamID]; exists {
				continue
			}

			membership := &portainer.TeamMembership{
				UserID: userID,
				TeamID: teamID,
				Role:   portainer.TeamMember,
			}
			if err := tx.TeamMembership().Create(membership); err != nil {
				return fmt.Errorf("create membership for team %d: %w", teamID, err)
			}
		}

		return nil
	})
}

func resolveOAuthTeamMemberships(settings *portainer.OAuthSettings, claims map[string]any) (map[portainer.TeamID]struct{}, map[portainer.TeamID]struct{}, error) {
	desiredTeamIDs := map[portainer.TeamID]struct{}{}
	managedTeamIDs := map[portainer.TeamID]struct{}{}
	claimValues := oauthClaimValues(claims[settings.TeamMemberships.OAuthClaimName])

	for _, mapping := range settings.TeamMemberships.OAuthClaimMappings {
		managedTeamIDs[mapping.Team] = struct{}{}

		matcher, err := regexp.Compile(mapping.ClaimValRegex)
		if err != nil {
			return nil, nil, fmt.Errorf("compile OAuth claim mapping %q: %w", mapping.ClaimValRegex, err)
		}

		for _, claimValue := range claimValues {
			if matcher.MatchString(claimValue) {
				desiredTeamIDs[mapping.Team] = struct{}{}
				break
			}
		}
	}

	if settings.DefaultTeamID != 0 {
		managedTeamIDs[settings.DefaultTeamID] = struct{}{}
		if len(desiredTeamIDs) == 0 {
			desiredTeamIDs[settings.DefaultTeamID] = struct{}{}
		}
	}

	return desiredTeamIDs, managedTeamIDs, nil
}

func oauthClaimValues(claim any) []string {
	switch values := claim.(type) {
	case string:
		return []string{values}
	case []string:
		return values
	case []any:
		result := make([]string, 0, len(values))
		for _, value := range values {
			if stringValue, ok := value.(string); ok {
				result = append(result, stringValue)
			}
		}
		return result
	default:
		return nil
	}
}
