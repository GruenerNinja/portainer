import {
  ResourceControlId,
  ResourceControlOwnership,
  ResourceControlResponse,
  ResourceControlType,
  ResourceId,
  TeamResourceAccess,
  UserResourceAccess,
  isReadWriteResourceAccess,
} from '../types';

export class ResourceControlViewModel {
  Id: ResourceControlId;

  Type: ResourceControlType;

  ResourceId: ResourceId;

  UserAccesses: UserResourceAccess[];

  TeamAccesses: TeamResourceAccess[];

  Public: boolean;

  System: boolean;

  Ownership: ResourceControlOwnership;

  constructor(data: ResourceControlResponse) {
    this.Id = data.Id;
    this.Type = data.Type;
    this.ResourceId = data.ResourceId;
    this.UserAccesses = data.UserAccesses;
    this.TeamAccesses = data.TeamAccesses;
    this.Public = data.Public;
    this.System = data.System;
    this.Ownership = determineOwnership(data);
  }
}

export function determineOwnership(resourceControl: ResourceControlResponse) {
  if (resourceControl.Public) {
    return ResourceControlOwnership.PUBLIC;
  }

  const readWriteUsers = resourceControl.UserAccesses.filter((access) =>
    isReadWriteResourceAccess(access.AccessLevel)
  );
  const readWriteTeams = resourceControl.TeamAccesses.filter((access) =>
    isReadWriteResourceAccess(access.AccessLevel)
  );

  if (readWriteUsers.length === 1 && readWriteTeams.length === 0) {
    return ResourceControlOwnership.PRIVATE;
  }

  if (readWriteUsers.length > 1 || readWriteTeams.length > 0) {
    return ResourceControlOwnership.RESTRICTED;
  }

  return ResourceControlOwnership.ADMINISTRATORS;
}
