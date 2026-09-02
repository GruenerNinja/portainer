import { buildVaultTokenCreateUrl } from './VaultTokenCreateLink';

describe('buildVaultTokenCreateUrl', () => {
  it('builds the token creation URL for a Vault server', () => {
    expect(buildVaultTokenCreateUrl('https://vault.example.com')).toBe(
      'https://vault.example.com/ui/vault/access/tokens/create'
    );
  });

  it('preserves a reverse-proxy base path', () => {
    expect(buildVaultTokenCreateUrl('https://example.com/vault/')).toBe(
      'https://example.com/vault/ui/vault/access/tokens/create'
    );
  });

  it('removes query parameters and fragments', () => {
    expect(
      buildVaultTokenCreateUrl('http://192.168.4.141:8200/?from=portainer#x')
    ).toBe('http://192.168.4.141:8200/ui/vault/access/tokens/create');
  });

  it.each(['', 'vault.internal', 'javascript:alert(1)'])(
    'rejects an invalid or unsafe address: %s',
    (address) => {
      expect(buildVaultTokenCreateUrl(address)).toBeUndefined();
    }
  );
});
