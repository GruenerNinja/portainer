import { useQuery } from '@tanstack/react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios/axios';

import { RbacRole } from './types';

export const roleQueryKeys = {
  all: ['roles'] as const,
};

export function useRbacRoles<T = Array<RbacRole>>({
  select,
}: {
  select?: (roles: Array<RbacRole>) => T;
} = {}) {
  return useQuery({
    select,
    queryKey: roleQueryKeys.all,
    queryFn: async () => {
      try {
        const { data } = await axios.get<Array<RbacRole>>('/roles');

        return data;
      } catch (e) {
        throw parseAxiosError(e, 'Failed to fetch roles');
      }
    },
  });
}
