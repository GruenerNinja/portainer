import { LockIcon } from 'lucide-react';
import { useFormikContext } from 'formik';

import { Card } from '@@/primitives/Card';
import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';

import { GitAuthentication } from '../../../components/GitAuthentication';

import { SettingsFormValues } from './types';

export function EditAuthWidget() {
  const { values, errors, setValues } = useFormikContext<SettingsFormValues>();
  const isVault = values.type === 'vault';

  return (
    <Card.Container>
      <Card.Header
        icon={LockIcon}
        title="Authentication"
        subtitle="Choose how Portainer authenticates to this source"
      />
      <Card.Body>
        {isVault ? (
          <FormControl
            inputId="vault-token"
            label="Token"
            errors={errors.token}
            tooltip="Leave empty to keep the saved token"
          >
            <Input
              id="vault-token"
              type="password"
              value={values.token}
              onChange={(e) =>
                setValues((oldValues) => ({
                  ...oldValues,
                  token: e.target.value,
                }))
              }
              data-cy="source-vault-token-input"
            />
          </FormControl>
        ) : (
          <GitAuthentication
            values={{
              authEnabled: values.authEnabled,
              username: values.username,
              password: values.password,
            }}
            isEditing
            errors={{ username: errors.username, password: errors.password }}
            onChange={(changed) =>
              setValues((oldValues) => ({ ...oldValues, ...changed }))
            }
            toggleDataCy="source-auth-enabled"
          />
        )}
      </Card.Body>
    </Card.Container>
  );
}
