import { useQuery } from '@tanstack/react-query';

import axios from '@/portainer/services/axios/axios';
import { withError } from '@/react-tools/react-query';
import { SourcesAutoUpdateInfo } from '@/react/portainer/generated-api/portainer/types.gen';

import { Source } from '../types';
import { Workflow } from '../../WorkflowsView/types';

import { sourceQueryKeys } from './query-keys';

export type ConnectionInfo = {
  configFilePath?: string;
  tlsSkipVerify?: boolean;
  authentication?: {
    username?: string;
  };
  vault?: {
    address: string;
    tlsSkipVerify: boolean;
    namespace?: string;
    kvVersion: number;
    authMethod: string;
  };
};
export type AutoUpdateInfo = SourcesAutoUpdateInfo;

export interface SourceDetail extends Source {
  connection: ConnectionInfo;
  autoUpdate?: AutoUpdateInfo;
  workflows: Workflow[];
}

async function getSource(id: Source['id']): Promise<SourceDetail> {
  const { data } = await axios.get<SourceDetail>(`/gitops/sources/${id}`);
  return data;
}

export function useSource(id: Source['id'] | undefined) {
  return useQuery({
    queryKey: sourceQueryKeys.detail(id!),
    queryFn: () => getSource(id!),
    enabled: !!id,
    ...withError('Failed loading source'),
  });
}
