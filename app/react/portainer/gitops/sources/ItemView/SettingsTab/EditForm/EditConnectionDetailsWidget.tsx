import { LinkIcon } from 'lucide-react';
import { useFormikContext } from 'formik';

import { Card } from '@@/primitives/Card';
import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { Select } from '@@/form-components/ReactSelect';
import { SwitchField } from '@@/form-components/SwitchField';

import { SettingsFormValues } from './types';

const kvVersionOptions = [
  { label: 'KV v2', value: 2 },
  { label: 'KV v1', value: 1 },
] as const;

export function EditConnectionDetailsWidget() {
  const { values, errors, setFieldValue } =
    useFormikContext<SettingsFormValues>();
  const isVault = values.type === 'vault';
  const selectedKVVersion =
    kvVersionOptions.find((option) => option.value === values.kvVersion) ??
    kvVersionOptions[0];

  return (
    <Card.Container>
      <Card.Header
        icon={LinkIcon}
        title="Connection Details"
        subtitle="Source name, URL, and connection settings"
      />
      <Card.Body>
        <FormControl inputId="name" label="Name" errors={errors.name} required>
          <Input
            id="name"
            name="name"
            value={values.name}
            onChange={(e) => setFieldValue('name', e.target.value)}
            data-cy="source-name-input"
          />
        </FormControl>
        <FormControl
          inputId="url"
          label={isVault ? 'Vault Address' : 'Repository URL'}
          errors={errors.url}
          required
        >
          <Input
            id="url"
            name="url"
            value={values.url}
            onChange={(e) => setFieldValue('url', e.target.value)}
            data-cy="source-url-input"
          />
        </FormControl>
        {isVault && (
          <>
            <FormControl
              inputId="namespace"
              label="Namespace"
              errors={errors.namespace}
            >
              <Input
                id="namespace"
                name="namespace"
                value={values.namespace}
                onChange={(e) => setFieldValue('namespace', e.target.value)}
                data-cy="source-vault-namespace-input"
              />
            </FormControl>
            <FormControl
              inputId="kvVersion"
              label="KV Engine"
              errors={errors.kvVersion}
              required
            >
              <Select
                inputId="kvVersion"
                data-cy="source-vault-kv-version"
                options={kvVersionOptions}
                value={selectedKVVersion}
                onChange={(option) =>
                  setFieldValue('kvVersion', option?.value ?? 2)
                }
              />
            </FormControl>
          </>
        )}
        <SwitchField
          label="Skip TLS verification"
          name="tlsSkipVerify"
          checked={values.tlsSkipVerify}
          onChange={(checked) => setFieldValue('tlsSkipVerify', checked)}
          data-cy="source-tls-skip-verify"
        />
      </Card.Body>
    </Card.Container>
  );
}
