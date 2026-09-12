import { useCallback } from 'react';
import { FormikErrors } from 'formik';

import { useCurrentUser } from '@/react/hooks/useUser';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { FormError } from '@@/form-components/FormError';

import { ResourceControlOwnership, AccessControlFormData } from '../types';

import { UsersField } from './UsersField';
import { TeamsField } from './TeamsField';
import { useLoadState } from './useLoadState';
import { AccessTypeSelector } from './AccessTypeSelector';

interface Props {
  values: AccessControlFormData;
  onChange(values: AccessControlFormData): void;
  isPublicVisible?: boolean;
  errors?: FormikErrors<AccessControlFormData>;
  formNamespace?: string;
  resourceName?: string;
  environmentId?: EnvironmentId;
  allowReadOnlyAccess?: boolean;
}

export function EditDetails({
  values,
  onChange,
  isPublicVisible = false,
  errors,
  formNamespace,
  resourceName = 'resource',
  environmentId,
  allowReadOnlyAccess = false,
}: Props) {
  const { user, isPureAdmin } = useCurrentUser();

  const { users, teams, isLoading } = useLoadState(environmentId);
  const readOnlyAuthorizedUsers = values.readOnlyAuthorizedUsers || [];
  const readOnlyAuthorizedTeams = values.readOnlyAuthorizedTeams || [];

  const handleChange = useCallback(
    (partialValues: Partial<typeof values>) => {
      const nextValues = { ...values, ...partialValues };
      nextValues.readOnlyAuthorizedUsers = (
        nextValues.readOnlyAuthorizedUsers || []
      ).filter((id) => !nextValues.authorizedUsers.includes(id));
      nextValues.readOnlyAuthorizedTeams = (
        nextValues.readOnlyAuthorizedTeams || []
      ).filter((id) => !nextValues.authorizedTeams.includes(id));
      onChange(nextValues);
    },

    [values, onChange]
  );

  if (
    isLoading ||
    !teams ||
    (isPureAdmin && !users) ||
    !values.authorizedUsers
  ) {
    return null;
  }

  return (
    <>
      <AccessTypeSelector
        onChange={handleChangeOwnership}
        name={withNamespace('ownership')}
        value={values.ownership}
        isAdmin={isPureAdmin}
        isPublicVisible={isPublicVisible}
        teams={teams}
        resourceName={resourceName}
      />

      {values.ownership === ResourceControlOwnership.RESTRICTED && (
        <div aria-label="extra-options">
          {isPureAdmin && (
            <UsersField
              name={withNamespace('authorizedUsers')}
              users={users || []}
              onChange={(authorizedUsers) => handleChange({ authorizedUsers })}
              value={values.authorizedUsers}
              errors={errors?.authorizedUsers}
            />
          )}

          {(isPureAdmin || teams.length > 1) && (
            <TeamsField
              name={withNamespace('authorizedTeams')}
              teams={teams}
              overrideTooltip={
                !isPureAdmin && teams.length > 1
                  ? 'As you are a member of multiple teams, you can select which teams(s) will be able to manage this resource.'
                  : undefined
              }
              onChange={(authorizedTeams) => handleChange({ authorizedTeams })}
              value={values.authorizedTeams}
              errors={errors?.authorizedTeams}
            />
          )}

          {typeof errors === 'string' && (
            <div className="form-group col-md-12">
              <FormError>{errors}</FormError>
            </div>
          )}
        </div>
      )}

      {allowReadOnlyAccess &&
        isPureAdmin &&
        [
          ResourceControlOwnership.PRIVATE,
          ResourceControlOwnership.RESTRICTED,
        ].includes(values.ownership) && (
          <div aria-label="read-only-options">
            <UsersField
              name={withNamespace('readOnlyAuthorizedUsers')}
              users={(users || []).filter(
                (candidate) => !values.authorizedUsers.includes(candidate.Id)
              )}
              onChange={(readOnlyAuthorizedUsers) =>
                handleChange({ readOnlyAuthorizedUsers })
              }
              value={readOnlyAuthorizedUsers}
              errors={errors?.readOnlyAuthorizedUsers}
              label="Read-only users"
              tooltip="These users can inspect this stack and read its configuration, but cannot update, redeploy, start, stop, or delete it."
              inputId="read-only-users-selector"
              dataCy="read-only-users-selector"
            />
            <TeamsField
              name={withNamespace('readOnlyAuthorizedTeams')}
              teams={teams.filter(
                (candidate) => !values.authorizedTeams.includes(candidate.Id)
              )}
              onChange={(readOnlyAuthorizedTeams) =>
                handleChange({ readOnlyAuthorizedTeams })
              }
              value={readOnlyAuthorizedTeams}
              errors={errors?.readOnlyAuthorizedTeams}
              label="Read-only teams"
              overrideTooltip="These teams can inspect this stack and read its configuration, but cannot update, redeploy, start, stop, or delete it."
              inputId="read-only-teams-selector"
              dataCy="read-only-teams-selector"
            />
          </div>
        )}
    </>
  );

  function withNamespace(name: string) {
    return formNamespace ? `${formNamespace}.${name}` : name;
  }

  function handleChangeOwnership(ownership: ResourceControlOwnership) {
    let { authorizedTeams, authorizedUsers } = values;

    if (ownership === ResourceControlOwnership.PRIVATE && user) {
      authorizedUsers = [user.Id];
      authorizedTeams = [];
    }

    if (ownership === ResourceControlOwnership.RESTRICTED) {
      authorizedUsers = [];
      authorizedTeams = [];
      // Non admin team leaders/members under only one team can
      // automatically grant the resource access to all members
      // under the team
      if (!isPureAdmin && teams && teams.length === 1) {
        authorizedTeams = teams.map((team) => team.Id);
      }
    }

    if (
      ownership === ResourceControlOwnership.PUBLIC ||
      ownership === ResourceControlOwnership.ADMINISTRATORS
    ) {
      handleChange({
        ownership,
        authorizedTeams,
        authorizedUsers,
        readOnlyAuthorizedTeams: [],
        readOnlyAuthorizedUsers: [],
      });
      return;
    }

    handleChange({ ownership, authorizedTeams, authorizedUsers });
  }
}
