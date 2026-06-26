import {
  object,
  array,
  number,
  mixed,
  SchemaOf,
  bool,
  string,
  ValidationError,
} from 'yup';

import { accessControlFormValidation } from '@/react/portainer/access-control/AccessControlForm';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { nameValidation } from '@/react/docker/stacks/common/NameField';
import { Stack } from '@/react/common/stacks/types';

import { envVarValidation } from '@@/form-components/EnvironmentVariablesFieldset';

import { BaseFormValues, FormValues } from './types';
import { getEditorValidationSchema } from './EditorSection/validation';
import { getGitValidationSchema } from './GitSection/validation';
import { getTemplateValidationSchema } from './TemplateSection/validation';
import { getUploadValidationSchema } from './UploadSection/validation';

export function getValidationSchema({
  isAdmin,
  environmentId,
  stacks,
  containerNames = [],
}: {
  isAdmin: boolean;
  environmentId: EnvironmentId;
  stacks?: Array<Stack>;
  containerNames?: Array<string>;
}): SchemaOf<FormValues> {
  return getBaseValidationSchema({ isAdmin, environmentId, stacks }).concat(
    object({
      git: getGitValidationSchema().when('method', {
        is: 'repository',
        then: (schema) => schema.required(),
        otherwise: () => mixed(),
      }),
      upload: getUploadValidationSchema({ containerNames }).when('method', {
        is: 'upload',
        then: (schema) => schema.required(),
        otherwise: () => mixed(),
      }),
      editor: getEditorValidationSchema({ containerNames }).when('method', {
        is: 'editor',
        then: (schema) => schema.required(),
        otherwise: () => mixed(),
      }),
      template: getTemplateValidationSchema({ containerNames }).when('method', {
        is: 'template',
        then: (schema) => schema.required(),
        otherwise: () => mixed(),
      }),
    })
  );
}

function getBaseValidationSchema({
  isAdmin,
  environmentId,
  stacks,
}: {
  isAdmin: boolean;
  environmentId: EnvironmentId;
  stacks?: Array<Stack>;
}): SchemaOf<BaseFormValues> {
  return object({
    method: mixed<'editor' | 'upload' | 'repository' | 'template'>()
      .oneOf(['editor', 'repository', 'template', 'upload'])
      .default('editor'),
    name: nameValidation({ environmentId, stacks }).required(
      'Stack name is required'
    ),
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
    ).test('unique', 'This secret key is already defined', (mappings, ctx) =>
      validateUniqueSecretKeys(mappings, ctx)
    ),
    accessControl: accessControlFormValidation(isAdmin),
    enableWebhook: bool().default(false),
    registries: array(number().required()).default([]),
  });
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
