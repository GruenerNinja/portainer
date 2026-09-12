import { useQuery } from '@tanstack/react-query';
import { useCurrentStateAndParams } from '@uirouter/react';

import { listRegistryCatalogs } from '@/react/portainer/registries/registry.service';
import { useRegistry } from '@/react/portainer/registries/queries/useRegistry';
import { queryKeys } from '@/react/portainer/registries/queries/query-keys';
import { withError } from '@/react-tools/react-query';

import { PageHeader } from '@@/PageHeader';

import { RepositoriesDatatable } from './RepositoriesDatatable';

export function RepositoriesView() {
  const { params } = useCurrentStateAndParams();
  const registryId = parseId(params.id);
  const registryQuery = useRegistry(registryId);
  const catalogQuery = useQuery({
    queryKey: [...queryKeys.item(registryId), 'catalog'],
    queryFn: () => listRegistryCatalogs(registryId),
    ...withError('Unable to retrieve registry repositories'),
  });

  const registryName = registryQuery.data?.Name || 'Registry';
  const repositories = catalogQuery.data?.repositories.map((Name) => ({
    Name,
  }));

  return (
    <>
      <PageHeader
        title={`${registryName} repositories`}
        breadcrumbs={[
          { label: 'Registries', link: 'portainer.registries' },
          registryName,
        ]}
        reload
      />
      <RepositoriesDatatable dataset={repositories} />
    </>
  );
}

function parseId(value: string | undefined) {
  const id = Number(value);
  if (!Number.isInteger(id) || id <= 0) {
    throw new Error('Missing registry id');
  }
  return id;
}
