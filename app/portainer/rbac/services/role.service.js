import axios from '@/portainer/services/axios/axios';

export function RoleService() {
  return {
    roles,
  };

  async function roles() {
    const { data } = await axios.get('/roles');
    return data;
  }
}
