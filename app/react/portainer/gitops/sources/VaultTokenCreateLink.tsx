import { ExternalLink } from '@@/ExternalLink';

interface Props {
  address: string;
}

export function VaultTokenCreateLink({ address }: Props) {
  const tokenCreateUrl = buildVaultTokenCreateUrl(address);
  if (!tokenCreateUrl) {
    return null;
  }

  return (
    <ExternalLink
      to={tokenCreateUrl}
      className="btn btn-default btn-sm no-link self-start"
      data-cy="vault-create-token-link"
    >
      Create token in Vault
    </ExternalLink>
  );
}

export function buildVaultTokenCreateUrl(address: string) {
  try {
    const url = new URL(address.trim());
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return undefined;
    }

    const basePath = url.pathname.replace(/\/+$/, '');
    url.pathname = `${basePath}/ui/vault/access/tokens/create`;
    url.search = '';
    url.hash = '';

    return url.toString();
  } catch {
    return undefined;
  }
}
