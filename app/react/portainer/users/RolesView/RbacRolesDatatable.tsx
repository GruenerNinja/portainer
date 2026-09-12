import { useMemo, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CellContext, createColumnHelper } from '@tanstack/react-table';
import { Edit, FileCode, Plus } from 'lucide-react';

import { promiseSequence } from '@/portainer/helpers/promise-utils';
import { notifySuccess } from '@/portainer/services/notifications';

import { Badge } from '@@/Badge';
import { Button } from '@@/buttons';
import { DeleteButton } from '@@/buttons/DeleteButton';
import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { RoleEditor } from './RoleEditor';
import {
  createRole,
  deleteRole,
  RolePayload,
  updateRole,
} from './role.service';
import { RbacRole } from './types';
import { roleQueryKeys, useRbacRoles } from './useRbacRoles';

const tableKey = 'rbac-roles-table';
const firstCustomRoleId = 6;
const store = createPersistedStore(tableKey);

export function RbacRolesDatatable() {
  const tableState = useTableState(store, tableKey);
  const rolesQuery = useRbacRoles({
    select: (roles) => [...roles].sort((a, b) => a.Priority - b.Priority),
  });
  const queryClient = useQueryClient();
  const [editingRole, setEditingRole] = useState<RbacRole | 'new'>();

  const saveMutation = useMutation(
    async ({
      role,
      payload,
    }: {
      role: RbacRole | 'new';
      payload: RolePayload;
    }) => (role === 'new' ? createRole(payload) : updateRole(role.Id, payload)),
    {
      meta: { error: { title: 'Failure', message: 'Unable to save role' } },
      onSuccess() {
        setEditingRole(undefined);
        notifySuccess('Success', 'Role saved');
        return queryClient.invalidateQueries(roleQueryKeys.all);
      },
    }
  );
  const deleteMutation = useMutation(
    async (roles: RbacRole[]) =>
      promiseSequence(roles.map((role) => () => deleteRole(role.Id))),
    {
      meta: { error: { title: 'Failure', message: 'Unable to remove role' } },
      onSuccess() {
        notifySuccess('Success', 'Roles removed');
        return queryClient.invalidateQueries(roleQueryKeys.all);
      },
    }
  );

  const columns = useMemo(() => getColumns(setEditingRole), []);
  const allAuthorizations = useMemo(
    () =>
      Array.from(
        new Set(
          (rolesQuery.data || []).flatMap((role) =>
            Object.keys(role.Authorizations || {})
          )
        )
      ).sort(),
    [rolesQuery.data]
  );

  return (
    <>
      <Datatable
        title="Roles"
        titleIcon={FileCode}
        dataset={rolesQuery.data || []}
        columns={columns}
        isLoading={rolesQuery.isLoading}
        settingsManager={tableState}
        isRowSelectable={({ original }) => original.Id >= firstCustomRoleId}
        renderTableActions={(selectedRoles) => (
          <>
            <DeleteButton
              disabled={!selectedRoles.length || deleteMutation.isLoading}
              onConfirmed={() => deleteMutation.mutate(selectedRoles)}
              confirmMessage="Delete the selected custom roles? Roles assigned to an access policy must be unassigned first."
              data-cy="remove-roles-button"
            />
            <Button
              icon={Plus}
              onClick={() => setEditingRole('new')}
              data-cy="add-role-button"
            >
              Add role
            </Button>
          </>
        )}
        data-cy="rbac-roles-datatable"
      />
      {editingRole && (
        <RoleEditor
          role={editingRole}
          availableAuthorizations={allAuthorizations}
          isSaving={saveMutation.isLoading}
          onDismiss={() => setEditingRole(undefined)}
          onSubmit={(payload) =>
            saveMutation.mutate({ role: editingRole, payload })
          }
        />
      )}
    </>
  );
}

function getColumns(onEdit: (role: RbacRole) => void) {
  const helper = createColumnHelper<RbacRole>();
  return [
    helper.accessor('Name', {
      header: 'Name',
      cell: ({ row, getValue }) => (
        <div className="flex items-center gap-2">
          <span>{getValue()}</span>
          {row.original.Id < firstCustomRoleId && (
            <Badge type="info">Built-in</Badge>
          )}
        </div>
      ),
    }),
    helper.accessor('Description', { header: 'Description' }),
    helper.accessor('Priority', { header: 'Priority' }),
    helper.display({
      id: 'actions',
      header: 'Actions',
      cell: ({ row: { original } }: CellContext<RbacRole, unknown>) =>
        original.Id >= firstCustomRoleId ? (
          <Button
            color="link"
            icon={Edit}
            onClick={() => onEdit(original)}
            data-cy={`edit-role-${original.Id}`}
          >
            Edit
          </Button>
        ) : null,
    }),
  ];
}
