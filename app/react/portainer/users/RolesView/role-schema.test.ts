import { zPortainerAccessPolicy } from '@api/zod.gen';

describe('access policy role identifiers', () => {
  it.each([0, 1, 5, 6, 42])('accepts role id %s', (RoleId) => {
    expect(zPortainerAccessPolicy.parse({ RoleId })).toEqual({ RoleId });
  });
});
