import { useQuery } from '@tanstack/react-query';

import { gitOpsSourcesTest } from '@api/sdk.gen';
import { type SourcesGitSourceCreatePayload } from '@api/types.gen';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { strToHash } from '@/react/utils/hash';

import { sourceQueryKeys } from '../queries/query-keys';

import { VaultSourcePayload } from './type';

type TestSourceConnectionPayload =
  | { type: 'git'; git: SourcesGitSourceCreatePayload }
  | { type: 'vault'; vault: VaultSourcePayload };

export function useTestSourceConnection(
  payload: TestSourceConnectionPayload | undefined
) {
  const payloadHashedPassword = {
    ...payload,
    git:
      payload?.type === 'git'
        ? {
            ...payload.git,
            authentication: {
              ...payload.git.authentication,
              password: null,
              passwordHash: payload.git.authentication?.password
                ? strToHash(payload.git.authentication.password)
                : undefined,
            },
          }
        : undefined,
    vault:
      payload?.type === 'vault'
        ? {
            ...payload.vault,
            authentication: {
              ...payload.vault.authentication,
              token: null,
              tokenHash: payload.vault.authentication.token
                ? strToHash(payload.vault.authentication.token)
                : undefined,
            },
          }
        : undefined,
  };

  return useQuery({
    queryKey: [
      ...sourceQueryKeys.all,
      'connection-test',
      payloadHashedPassword,
    ],
    queryFn: async () => {
      if (!payload) {
        throw new Error('Connection details are required');
      }
      if (payload.type === 'vault') {
        try {
          const { data } = await axios.post(
            '/gitops/sources/vault/test',
            payload.vault
          );
          return data;
        } catch (e) {
          throw parseAxiosError(e as Error);
        }
      }
      const { data } = await gitOpsSourcesTest({ body: payload.git });
      return data;
    },
    enabled: !!payload,
    retry: false,
    refetchOnWindowFocus: false,
    staleTime: 5000,
  });
}
