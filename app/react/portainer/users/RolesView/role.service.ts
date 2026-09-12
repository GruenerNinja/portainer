import axios, { parseAxiosError } from '@/portainer/services/axios/axios';

import { AuthorizationMap, RbacRole } from './types';

export interface RolePayload {
  Name: string;
  Description: string;
  Authorizations: AuthorizationMap;
  Priority: number;
}

export async function createRole(payload: RolePayload) {
  try {
    const { data } = await axios.post<RbacRole>('/roles', payload);
    return data;
  } catch (error) {
    throw parseAxiosError(error, 'Unable to create role');
  }
}

export async function updateRole(id: number, payload: RolePayload) {
  try {
    const { data } = await axios.put<RbacRole>(`/roles/${id}`, payload);
    return data;
  } catch (error) {
    throw parseAxiosError(error, 'Unable to update role');
  }
}

export async function deleteRole(id: number) {
  try {
    await axios.delete(`/roles/${id}`);
  } catch (error) {
    throw parseAxiosError(error, 'Unable to delete role');
  }
}
