import {
  type SourcesGitAuthenticationPayload,
  type SourcesGitSourceCreatePayload,
} from '@api/types.gen';

import {
  AccessControlFormData,
  ResourceControlOwnership,
} from '@/react/portainer/access-control/types';

import { CreateSourcePayload } from './useSourceCreateMutation';

type GitFormValues = {
  url: string;
  authentication: {
    authEnabled: boolean;
    username?: string;
    password?: string;
  };
  tlsSkipVerify?: boolean;
  polling: {
    enabled: boolean;
    interval: string;
  };
  /** Mirrors the connection-test result; not sent in the create payload. */
  connectionOk: boolean;
};

export type VaultFormValues = {
  address: string;
  namespace?: string;
  kvVersion: 1 | 2;
  tlsSkipVerify?: boolean;
  authentication: {
    method: 'token';
    token?: string;
  };
  /** Mirrors the connection-test result; not sent in the create payload. */
  connectionOk: boolean;
};

export type VaultSourcePayload = {
  name?: string;
  address: string;
  namespace?: string;
  kvVersion?: 1 | 2;
  tlsSkipVerify?: boolean;
  administratorsOnly?: boolean;
  public?: boolean;
  teamAccesses?: Array<number>;
  userAccesses?: Array<number>;
  authentication: {
    method: 'token';
    token: string;
  };
};

export const FormValueTypes = ['git', 'registry', 'helm', 'vault'] as const;

export type FormValues = AccessControlFormData & {
  name: string;
  type: (typeof FormValueTypes)[number];
  git: GitFormValues;
  vault: VaultFormValues;
};

export function formValuesToCreatePayload(
  values: FormValues & { type: 'git' }
): Extract<CreateSourcePayload, { type: 'git' }>;
export function formValuesToCreatePayload(values: FormValues): CreateSourcePayload;
export function formValuesToCreatePayload(values: FormValues): CreateSourcePayload {
  const { name, type, authorizedTeams, authorizedUsers, ownership } = values;
  const accessControl = {
    administratorsOnly: ownership === ResourceControlOwnership.ADMINISTRATORS,
    public: ownership === ResourceControlOwnership.PUBLIC,
    teamAccesses: authorizedTeams,
    userAccesses: authorizedUsers,
  };

  if (type === 'vault') {
    return {
      type,
      vault: {
        ...vaultFormValuesToPayload(name, values.vault),
        ...accessControl,
      },
    };
  }

  const { authentication, tlsSkipVerify, url, polling } = values.git;
  return {
    type: 'git',
    git: {
      name,
      tlsSkipVerify,
      url,
      authentication: buildAuthPayload(authentication),
      ...accessControl,
      interval: polling.enabled ? polling.interval : '',
    },
  };
}

export function gitFormValuesToTestPayload({
  authentication,
  url,
  tlsSkipVerify,
}: GitFormValues): SourcesGitSourceCreatePayload {
  return {
    url,
    tlsSkipVerify,
    authentication: buildAuthPayload(authentication),
  };
}

export function vaultFormValuesToTestPayload(
  values: VaultFormValues
): VaultSourcePayload {
  return vaultFormValuesToPayload('', values);
}

function vaultFormValuesToPayload(
  name: string,
  {
    address,
    namespace,
    kvVersion,
    tlsSkipVerify,
    authentication,
  }: VaultFormValues
): VaultSourcePayload {
  return {
    name,
    address,
    namespace,
    kvVersion,
    tlsSkipVerify,
    authentication: {
      method: authentication.method,
      token: authentication.token || '',
    },
  };
}

function buildAuthPayload(
  auth: GitFormValues['authentication']
): SourcesGitAuthenticationPayload | undefined {
  const { authEnabled, username, password } = auth;
  if (!authEnabled || !username || !password) {
    return undefined;
  }
  return { username, password };
}
