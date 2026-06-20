import { bool, mixed, object, string } from 'yup';

import { isValidUrl } from '@@/form-components/validate-url';

import { FormValues, FormValueTypes } from './type';

export function validationSchema() {
  return object({
    name: string().required('Name is required.'),
    type: mixed<FormValues['type']>()
      .oneOf([...FormValueTypes])
      .required()
      .default('git'),
    git: mixed().when('type', {
      is: 'git',
      then: validateGit(),
      otherwise: mixed().notRequired(),
    }),
    vault: mixed().when('type', {
      is: 'vault',
      then: validateVault(),
      otherwise: mixed().notRequired(),
    }),
  });
}

export function validateGitConnection() {
  return validateGit().pick(['url', 'authentication', 'tlsSkipVerify']);
}

export function validateVaultConnection() {
  return validateVault().pick([
    'address',
    'namespace',
    'kvVersion',
    'tlsSkipVerify',
    'authentication',
  ]);
}

function validateGit() {
  return object({
    authentication: object({
      authEnabled: bool().required().default(false),
      username: string().when('authEnabled', {
        is: true,
        then: string().required('Username is required'),
      }),
      password: string().when('authEnabled', {
        is: true,
        then: string().required('Password is required'),
      }),
    }),
    url: string()
      .required('Repository URL is required.')
      .test(
        'valid repository URL',
        'The repository URL must be a valid URL (localhost cannot be used)',
        (value) =>
          isValidUrl(
            value,
            (url) => !!url.hostname && url.hostname !== 'localhost'
          )
      ),
    tlsSkipVerify: bool(),
    connectionOk: bool()
      .oneOf([true], 'The connection test must succeed before continuing.')
      .required(),
  });
}

function validateVault() {
  return object({
    address: string()
      .required('Vault address is required.')
      .test(
        'valid vault address',
        'The Vault address must be a valid URL (localhost cannot be used)',
        (value) =>
          isValidUrl(
            value,
            (url) => !!url.hostname && url.hostname !== 'localhost'
          )
      ),
    namespace: string().optional(),
    kvVersion: mixed<1 | 2>().oneOf([1, 2]).required('KV version is required.'),
    tlsSkipVerify: bool(),
    authentication: object({
      method: mixed<'token'>().oneOf(['token']).required(),
      token: string().required('Token is required.'),
    }),
    connectionOk: bool()
      .oneOf([true], 'The connection test must succeed before continuing.')
      .required(),
  });
}
