import { useMutation } from '@tanstack/react-query';

import {
  type SourcesConnectionTestResult,
  type SourcesGitSourceUpdatePayload,
} from '@api/types.gen';
import { gitOpsSourcesTestById } from '@api/sdk.gen';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';
import { withError } from '@/react-tools/react-query';

import { Source } from '../types';

import { UpdateSourcePayload } from './useUpdateSourceMutation';

export type ConnectionTestResult = SourcesConnectionTestResult;

async function testSourceConnection(
  id: Source['id'],
  payload: UpdateSourcePayload
): Promise<ConnectionTestResult> {
  if (payload.type === 'vault') {
    const { type, ...body } = payload;
    try {
      const { data } = await axios.post<ConnectionTestResult>(
        `/gitops/sources/${id}/test`,
        body
      );
      return data;
    } catch (e) {
      throw parseAxiosError(e as Error);
    }
  }

  const { type, ...body } = payload as SourcesGitSourceUpdatePayload & {
    type?: 'git';
  };
  const { data } = await gitOpsSourcesTestById({ path: { id }, body });
  return data;
}

export function useTestSourceConnectionMutation() {
  return useMutation({
    mutationFn: ({
      id,
      payload = {},
    }: {
      id: Source['id'];
      payload?: UpdateSourcePayload;
    }) => testSourceConnection(id, payload),
    ...withError('Connection test failed'),
  });
}
