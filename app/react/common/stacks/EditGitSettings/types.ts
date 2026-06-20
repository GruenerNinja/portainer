import { GitFormModel } from '@/react/portainer/gitops/types';
import { StackSecretMapping } from '@/react/common/stacks/types';

import { EnvVarValues } from '@@/form-components/EnvironmentVariablesFieldset';

export interface FormValues {
  kube: { name: string };
  git: GitFormModel;
  env: EnvVarValues;
  secretMappings: StackSecretMapping[];
  prune: boolean;
  redeployNow: boolean;
}
