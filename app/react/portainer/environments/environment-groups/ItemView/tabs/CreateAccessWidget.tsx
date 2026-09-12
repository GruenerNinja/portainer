import { useState } from 'react';
import { UserPlus, Plus } from 'lucide-react';

import { RoleTypes } from '@/portainer/rbac/models/role';
import { useRbacRoles } from '@/react/portainer/users/RolesView/useRbacRoles';
import {
  Option,
  PorAccessManagementUsersSelector,
} from '@/react/portainer/access-control/AccessManagement/PorAccessManagementUsersSelector';

import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';
import { TextTip } from '@@/Tip/TextTip';
import { LoadingButton } from '@@/buttons';
import { FormControl } from '@@/form-components/FormControl';
import { PortainerSelect } from '@@/form-components/PortainerSelect';

interface Props {
  availableUsersAndTeams: Array<Option>;
  isLoading: boolean;
  isUpdating: boolean;
  onSubmit(
    usersAndTeams: Array<Option>,
    roleId: number,
    onSuccess: () => void
  ): void;
}

export function CreateAccessWidget({
  availableUsersAndTeams,
  isLoading,
  isUpdating,
  onSubmit,
}: Props) {
  const [selectedUsersAndTeams, setSelectedUsersAndTeams] = useState<
    Array<Option>
  >([]);
  const [selectedRoleId, setSelectedRoleId] = useState<number>(
    RoleTypes.STANDARD
  );

  const rolesQuery = useRbacRoles();
  const roleOptions = (rolesQuery.data || []).map((role) => ({
    label: role.Name,
    value: role.Id,
  }));

  return (
    <Widget aria-label="Create access">
      <WidgetTitle icon={UserPlus} title="Create access" />
      <WidgetBody>
        <TextTip className="mb-4" childrenWrapperClassName="text-warning">
          Adding user access will require the affected user(s) to logout and
          login for the changes to be taken into account.
        </TextTip>

        <form className="form-horizontal" onSubmit={handleSubmit}>
          <PorAccessManagementUsersSelector
            options={availableUsersAndTeams}
            value={selectedUsersAndTeams}
            onChange={(value) => setSelectedUsersAndTeams([...value])}
            isLoading={isLoading}
          />

          <FormControl label="Role" inputId="role-selector">
            <PortainerSelect
              inputId="role-selector"
              value={selectedRoleId}
              onChange={(roleId) =>
                setSelectedRoleId(roleId ?? RoleTypes.STANDARD)
              }
              options={roleOptions}
              disabled={rolesQuery.isLoading}
              data-cy="access-management-role-select"
            />
          </FormControl>

          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                type="submit"
                className="!ml-0"
                disabled={selectedUsersAndTeams.length === 0}
                isLoading={isUpdating}
                loadingText="Creating access..."
                icon={Plus}
                data-cy="access-createAccess"
              >
                Create access
              </LoadingButton>
            </div>
          </div>
        </form>
      </WidgetBody>
    </Widget>
  );

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSubmit(selectedUsersAndTeams, selectedRoleId, () =>
      setSelectedUsersAndTeams([])
    );
  }
}
