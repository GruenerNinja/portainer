import {
  type LucideIcon,
  Box,
  GitBranch,
  KeyRound,
  Package,
} from 'lucide-react';

import { WorkflowsStatus } from '@api/types.gen';

export type SourceStatus = WorkflowsStatus;
export type SourceType = 'git' | 'helm' | 'oci' | 'vault';

export interface Source {
  id: number;
  name: string;
  type: SourceType;
  url: string;
  status: SourceStatus;
  error?: string;
  usedBy: number;
  environments: number;
  lastSync: number;
}

export const SOURCE_TYPES: Record<
  SourceType,
  { label: string; icon: LucideIcon }
> = {
  git: { label: 'Git', icon: GitBranch },
  helm: { label: 'Helm', icon: Package },
  oci: { label: 'OCI', icon: Box },
  vault: { label: 'Vault', icon: KeyRound },
};
