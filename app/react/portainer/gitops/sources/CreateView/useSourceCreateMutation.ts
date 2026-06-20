import { useMutation, useQueryClient } from '@tanstack/react-query';

import { gitOpsSourcesCreateGit } from '@api/sdk.gen';
import { type SourcesGitSourceCreatePayload } from '@api/types.gen';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { withError, withInvalidate } from '@/react-tools/react-query';

import { sourceQueryKeys } from '../queries/query-keys';

import { VaultSourcePayload } from './type';

export type CreateSourcePayload =
  | {
      type: 'git';
      git: SourcesGitSourceCreatePayload;
    }
  | {
      type: 'vault';
      vault: VaultSourcePayload;
    };

export function useCreateSourceMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createSource,
    ...withError('Unable to create source'),
    ...withInvalidate(queryClient, [sourceQueryKeys.all]),
  });
}

async function createSource(payload: CreateSourcePayload) {
  if (payload.type === 'vault') {
    try {
      const { data } = await axios.post('/gitops/sources/vault', payload.vault);
      return data;
    } catch (e) {
      throw parseAxiosError(e as Error);
    }
  }

  const { data } = await gitOpsSourcesCreateGit({ body: payload.git });
  return data;
}
