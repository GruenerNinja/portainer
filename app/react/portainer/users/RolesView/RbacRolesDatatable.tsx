import { FileCode } from 'lucide-react';
import { createColumnHelper } from '@tanstack/react-table';

import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { RbacRole } from './types';

const tableKey = 'rbac-roles-table';

const store = createPersistedStore(tableKey);

const columns = getColumns();

export function RbacRolesDatatable({
  dataset,
}: {
  dataset: Array<RbacRole> | undefined;
}) {
  const tableState = useTableState(store, tableKey);

  return (
    <Datatable
      title="Roles"
      titleIcon={FileCode}
      dataset={dataset || []}
      columns={columns}
      isLoading={!dataset}
      settingsManager={tableState}
      disableSelect
      data-cy="rbac-roles-datatable"
    />
  );
}

function getColumns() {
  const columnHelper = createColumnHelper<RbacRole>();

  return [
    columnHelper.accessor('Name', {
      header: 'Name',
    }),
    columnHelper.accessor('Description', {
      header: 'Description',
    }),
  ];
}
