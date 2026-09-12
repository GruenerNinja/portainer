import { useQuery } from '@tanstack/react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { withError } from '@/react-tools/react-query';

import { Registry } from '../types/registry';

import { buildUrl } from './build-url';
import { queryKeys } from './query-keys';

export function useRegistry(
  registryId?: Registry['Id'],
  shouldShowError: boolean = true
) {
  // Registry details are also available from the global administrator view,
  // where there is no environment in the route.
  const environmentId = useEnvironmentId(false);

  return useQuery(
    registryId ? queryKeys.item(registryId) : [],
    () => (registryId ? getRegistry(registryId, environmentId) : undefined),
    {
      enabled: !!registryId,
      retry: 1,
      ...(shouldShowError ? withError('Unable to load registry details') : {}),
    }
  );
}

async function getRegistry(registryId: Registry['Id'], environmentId?: number) {
  try {
    const { data } = await axios.get<Registry>(buildUrl(registryId), {
      params: {
        endpointId: environmentId || undefined,
      },
    });
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to retrieve registry');
  }
}
