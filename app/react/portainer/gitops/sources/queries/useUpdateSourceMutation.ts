import { useMutation, useQueryClient } from '@tanstack/react-query';

import { gitOpsSourcesUpdateGit } from '@api/sdk.gen';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { withError } from '@/react-tools/react-query';
import { SourcesGitSourceUpdatePayload } from '@/react/portainer/generated-api/portainer/types.gen';

import { Source } from '../types';

import { sourceQueryKeys } from './query-keys';

export type VaultSourceUpdatePayload = {
  name?: string;
  address?: string;
  tlsSkipVerify?: boolean;
  namespace?: string;
  kvVersion?: number;
  authentication?: {
    method?: 'token';
    token?: string;
  };
};

export type UpdateSourcePayload =
  | (SourcesGitSourceUpdatePayload & { type?: 'git' })
  | (VaultSourceUpdatePayload & { type?: 'vault' });

async function updateSource(
  id: Source['id'],
  payload: UpdateSourcePayload
): Promise<void> {
  if (payload.type === 'vault') {
    const { type, ...body } = payload;
    try {
      await axios.put(`/gitops/sources/${id}`, body);
      return;
    } catch (e) {
      throw parseAxiosError(e as Error);
    }
  }

  const { type, ...body } = payload as SourcesGitSourceUpdatePayload & {
    type?: 'git';
  };
  await gitOpsSourcesUpdateGit({ path: { id }, body });
}

export function useUpdateSourceMutation(id: Source['id']) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: UpdateSourcePayload) => updateSource(id, payload),
    onSuccess: () => {
      return queryClient.invalidateQueries({
        queryKey: sourceQueryKeys.detail(id),
      });
    },
    ...withError('Unable to update source'),
  });
}
