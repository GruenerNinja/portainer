import { AuthTypeOption } from '@/react/portainer/account/git-credentials/types';

import {
  AutoUpdateModel,
  getDefaultAutoUpdateValues,
} from './AutoUpdateFieldset/utils';

export type {
  AutoUpdateMechanism,
  AutoUpdateModel,
  AutoUpdateResponse,
} from './AutoUpdateFieldset/utils';

export { type RelativePathModel } from './RelativePathFieldset/types';

export interface GitAuthenticationResponse {
  Username?: string;
  Password?: string;
  AuthorizationType?: AuthTypeOption;
}

export interface RepoConfigResponse {
  URL: string;
  ReferenceName: string;
  ConfigFilePath: string;
  Authentication?: GitAuthenticationResponse;
  ConfigHash: string;
  TLSSkipVerify: boolean;
}

export type GitAuthModel = {
  RepositoryAuthentication?: boolean;
  RepositoryUsername?: string;
  RepositoryPassword?: string;
  RepositoryAuthorizationType?: AuthTypeOption;
};

export type DeployMethod = 'compose' | 'manifest' | 'helm';

export interface GitFormModel extends GitAuthModel {
  RepositoryURL: string;
  RepositoryURLValid?: boolean;
  ComposeFilePathInRepository?: string;
  RepositoryReferenceName?: string;
  AdditionalFiles?: string[];
  TLSSkipVerify?: boolean;
  /**
   * Auto update
   *
   * if undefined, GitForm won't show the AutoUpdate fieldset
   */
  AutoUpdate?: AutoUpdateModel;
  /** ID of an existing Source. When set, inline URL and credentials are ignored. */
  SourceId?: number;
  SupportRelativePath?: boolean;
  FilesystemPath?: string;
}

export function getDefaultModel(
  autoUpdate: AutoUpdateModel = getDefaultAutoUpdateValues()
): GitFormModel {
  return {
    RepositoryURL: '',
    ComposeFilePathInRepository: 'docker-compose.yml',
    RepositoryReferenceName: 'refs/heads/main',
    RepositoryAuthentication: false,
    TLSSkipVerify: false,
    AutoUpdate: autoUpdate,
  };
}

export function toGitFormModel(
  response?: RepoConfigResponse,
  autoUpdate?: AutoUpdateModel
): GitFormModel {
  if (!response) {
    return getDefaultModel(autoUpdate);
  }

  const { URL, ReferenceName, ConfigFilePath, Authentication, TLSSkipVerify } =
    response;

  return {
    RepositoryURL: URL,
    ComposeFilePathInRepository: ConfigFilePath,
    RepositoryReferenceName: ReferenceName,
    RepositoryAuthentication: !!Authentication?.Username,
    RepositoryUsername: Authentication?.Username,
    RepositoryPassword: Authentication?.Password,
    RepositoryAuthorizationType: Authentication?.AuthorizationType,
    TLSSkipVerify,
    AutoUpdate: autoUpdate,
  };
}
