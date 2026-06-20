import { useMemo } from 'react';
import { array, boolean, number, object, SchemaOf, string } from 'yup';

import { StackType } from '@/react/common/stacks/types';
import { buildGitValidationSchema } from '@/react/portainer/gitops/GitForm';

import { envVarValidation } from '@@/form-components/EnvironmentVariablesFieldset';
import { buildUniquenessTest } from '@@/form-components/validate-unique';

import { FormValues } from './types';

export function useValidationSchema(
  stackType: StackType,
  isSourceSelection: boolean
): SchemaOf<FormValues> {
  const isKubernetes = stackType === StackType.Kubernetes;

  return useMemo(
    () =>
      object({
        kube: isKubernetes
          ? object({
              name: string().default(''),
            }).required()
          : object({ name: string().default('') }).optional(),
        git: buildGitValidationSchema(
          false,
          isKubernetes ? 'manifest' : 'compose',
          true,
          isSourceSelection
        ),

        env: envVarValidation(),
        secretMappings: array(
          object({
            name: string().default(''),
            sourceId: number()
              .min(1, 'Vault provider is required')
              .required('Vault provider is required'),
            path: string().required('Secret path is required'),
            key: string().required('Secret key is required'),
          })
        ).test(
          'unique',
          'This secret key is already defined',
          buildUniquenessTest(() => 'This secret key is already defined', 'key')
        ),
        prune: boolean().default(false),
        redeployNow: boolean().default(false),
      }),
    [isKubernetes, isSourceSelection]
  );
}
