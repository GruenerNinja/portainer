import { useQuery } from '@tanstack/react-query';

import {
  type SourcesConnectionInfo,
  type SourcesSourceDetail,
} from '@api/types.gen';
import { gitOpsSourceGet } from '@api/sdk.gen';

import { withError } from '@/react-tools/react-query';

import { Source } from '../types';

import { sourceQueryKeys } from './query-keys';

export type SourceDetail = Omit<SourcesSourceDetail, 'connection' | 'type'> & {
  connection: SourcesConnectionInfo & {
    vault?: {
      address: string;
      tlsSkipVerify: boolean;
      namespace?: string;
      kvVersion: number;
      authMethod: string;
    };
  };
  type: Source['type'];
};

export function sourceOptions(id: Source['id']) {
  return {
    queryKey: sourceQueryKeys.detail(id!),
    queryFn: () => getSource(id!),
    ...withError('Failed loading source'),
  };
}

export function useSource(id: Source['id'] | undefined) {
  return useQuery({
    ...sourceOptions(id!),
    enabled: !!id,
  });
}

export async function getSource(id: Source['id']): Promise<SourceDetail> {
  const { data } = await gitOpsSourceGet({ path: { id } });

  return data as SourceDetail;
}
