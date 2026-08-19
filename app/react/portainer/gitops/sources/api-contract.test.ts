import {
  zGitOpsSourcesListQuery,
  zGitOpsSourcesListResponse,
  zSourcesSourceDetail,
} from '@api/zod.gen';

describe('Vault source API contract', () => {
  it('accepts Vault in list filters and responses', () => {
    expect(zGitOpsSourcesListQuery.parse({ type: 'vault' })).toEqual({
      type: 'vault',
    });

    expect(
      zGitOpsSourcesListResponse.parse([
        {
          id: 1,
          name: 'production-vault',
          status: 'healthy',
          type: 'vault',
          url: 'https://vault.example.com',
        },
      ])
    ).toHaveLength(1);
  });

  it('accepts a Vault source detail response', () => {
    expect(
      zSourcesSourceDetail.parse({
        connection: {
          vault: {
            address: 'https://vault.example.com',
            authMethod: 'token',
            kvVersion: 2,
            tlsSkipVerify: false,
          },
        },
        id: 1,
        name: 'production-vault',
        status: 'healthy',
        type: 'vault',
        url: 'https://vault.example.com',
      }).type
    ).toBe('vault');
  });
});
