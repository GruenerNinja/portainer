import { useCurrentStateAndParams } from '@uirouter/react';

import { useRegistry } from '@/react/portainer/registries/queries/useRegistry';

import { PageHeader } from '@@/PageHeader';

import { useRepositoryTags } from '../ListView/useRepositoryTags';

import { TagsDatatable } from './TagsDatatable/TagsDatatable';

export function RepositoryView() {
  const { params } = useCurrentStateAndParams();
  const registryId = parseId(params.id);
  const environmentId = parseOptionalId(params.endpointId);
  const repository = parseRepository(params.repository);
  const registryQuery = useRegistry(registryId);
  const tagsQuery = useRepositoryTags({
    registryId,
    environmentId,
    repository,
  });
  const registryName = registryQuery.data?.Name || 'Registry';
  const tags = tagsQuery.data?.tags?.map((Name) => ({ Name })) || [];

  return (
    <>
      <PageHeader
        title={repository}
        breadcrumbs={[
          { label: 'Registries', link: 'portainer.registries' },
          {
            label: registryName,
            link: 'portainer.registries.registry.repositories',
            linkParams: { id: registryId, endpointId: environmentId },
          },
          repository,
        ]}
        reload
      />
      <TagsDatatable
        dataset={tagsQuery.isLoading ? undefined : tags}
        advancedFeaturesAvailable={false}
        onRemove={() => {}}
        onRetag={async () => {}}
      />
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

function parseOptionalId(value: string | undefined) {
  const id = Number(value);
  return Number.isInteger(id) && id > 0 ? id : undefined;
}

function parseRepository(value: string | undefined) {
  if (!value) {
    throw new Error('Missing repository name');
  }
  return value;
}
