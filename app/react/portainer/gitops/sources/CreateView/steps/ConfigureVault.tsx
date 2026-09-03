import { useFormikContext } from 'formik';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { Select } from '@@/form-components/ReactSelect';
import { SwitchField } from '@@/form-components/SwitchField';

import { VaultTokenCreateLink } from '../../VaultTokenCreateLink';
import { FormValues } from '../type';

import { ConnectionTest } from './ConnectionTest';

const kvVersionOptions = [
  { label: 'KV v2', value: 2 },
  { label: 'KV v1', value: 1 },
] as const;

export function ConfigureVault() {
  const { values, setFieldValue, errors } = useFormikContext<FormValues>();

  if (values.type !== 'vault') {
    return null;
  }

  const selectedKVVersion =
    kvVersionOptions.find(
      (option) => option.value === values.vault.kvVersion
    ) ?? kvVersionOptions[0];

  return (
    <div className="grid">
      <FormControl
        inputId="vault-address-input"
        label="Public Vault Address"
        required
        errors={errors.vault?.address}
        tooltip="Enter the stable domain URL of the Vault server, for example https://vault.example.com. Do not paste a Vault UI secret URL here."
      >
        <Input
          id="vault-address-input"
          value={values.vault.address}
          data-cy="vault-address-input"
          placeholder="https://vault.example.com"
          required
          onChange={({ target: { value } }) =>
            setFieldValue('vault.address', value)
          }
        />
      </FormControl>

      <FormControl
        inputId="vault-internal-address-input"
        label="Internal Vault Address"
        errors={errors.vault?.internalAddress}
        tooltip="Optional direct URL reachable from Portainer, for example http://vault:8200 or http://192.168.4.20:8200. Portainer uses it first and falls back to the public address, so Vault access does not depend on the public reverse proxy."
      >
        <Input
          id="vault-internal-address-input"
          value={values.vault.internalAddress}
          data-cy="vault-internal-address-input"
          placeholder="http://vault:8200"
          onChange={({ target: { value } }) =>
            setFieldValue('vault.internalAddress', value)
          }
        />
      </FormControl>

      <FormControl
        inputId="vault-namespace-input"
        label="Namespace"
        errors={errors.vault?.namespace}
        tooltip="Only for Vault Enterprise or HCP namespaces. Leave empty for normal Vault; this is not the KV mount or secret folder."
      >
        <Input
          id="vault-namespace-input"
          value={values.vault.namespace}
          data-cy="vault-namespace-input"
          placeholder="admin/team-a"
          onChange={({ target: { value } }) =>
            setFieldValue('vault.namespace', value)
          }
        />
      </FormControl>

      <FormControl
        inputId="vault-kv-version"
        label="KV engine version"
        required
        errors={errors.vault?.kvVersion}
      >
        <Select
          inputId="vault-kv-version"
          data-cy="vault-kv-version"
          options={kvVersionOptions}
          value={selectedKVVersion}
          onChange={(option) =>
            setFieldValue('vault.kvVersion', option?.value ?? 2)
          }
        />
      </FormControl>

      <SwitchField
        label="Skip TLS Verification"
        labelClass="col-sm-3 col-lg-2"
        name="VaultTLSSkipVerify"
        checked={values.vault.tlsSkipVerify || false}
        onChange={(value) => setFieldValue('vault.tlsSkipVerify', value)}
        tooltip="Enabling this will allow skipping TLS validation for any self-signed certificate."
        data-cy="vault-tls-skip-verify"
      />

      <FormControl
        inputId="vault-token-input"
        label="Token"
        required
        errors={errors.vault?.authentication?.token}
        tooltip="Vault token used by Portainer when resolving stack secrets. Renewable periodic tokens are checked hourly and renewed after half their period has elapsed."
      >
        <div className="flex flex-col gap-2">
          <Input
            id="vault-token-input"
            type="password"
            value={values.vault.authentication.token}
            data-cy="vault-token-input"
            required
            onChange={({ target: { value } }) =>
              setFieldValue('vault.authentication.token', value)
            }
          />
          <VaultTokenCreateLink address={values.vault.address} />
        </div>
      </FormControl>

      <ConnectionTest />
    </div>
  );
}
