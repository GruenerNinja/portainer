import { useMemo } from 'react';
import {
  array,
  boolean,
  number,
  object,
  SchemaOf,
  string,
  ValidationError,
} from 'yup';

import { StackType } from '@/react/common/stacks/types';
import { buildGitValidationSchema } from '@/react/portainer/gitops/GitForm';

import { envVarValidation } from '@@/form-components/EnvironmentVariablesFieldset';

import { FormValues } from './types';

export function useValidationSchema(
  stackType: StackType
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
        git: buildGitValidationSchema(isKubernetes ? 'manifest' : 'compose'),

        env: envVarValidation(),
        secretMappings: array(
          object({
            name: string().default(''),
            sourceId: number()
              .min(1, 'Vault provider is required')
              .required('Vault provider is required'),
            path: string().required('Secret path is required'),
            key: string().default(''),
          })
        ).test(
          'unique',
          'This secret key is already defined',
          (mappings, ctx) => validateUniqueSecretKeys(mappings, ctx)
        ),
        prune: boolean().default(false),
        redeployNow: boolean().default(false),
      }),
    [isKubernetes]
  );
}

function validateUniqueSecretKeys(
  mappings: Array<{ key?: string }> | undefined,
  ctx: {
    path: string;
    createError(args: { path: string; message: string }): ValidationError;
  }
) {
  const seen = new Set<string>();

  for (const [index, mapping] of (mappings ?? []).entries()) {
    const key = mapping.key?.trim();
    if (!key) {
      continue;
    }

    if (seen.has(key)) {
      return ctx.createError({
        path: `${ctx.path}[${index}].key`,
        message: 'This secret key is already defined',
      });
    }

    seen.add(key);
  }

  return true;
}
