import { useEffect } from 'react';
import { useFormikContext } from 'formik';
import { isEqual } from 'lodash';

import { useDebouncedValue } from '@/react/hooks/useDebouncedValue';

import { Alert } from '@@/Alert';

import {
  FormValues,
  gitFormValuesToTestPayload,
  vaultFormValuesToTestPayload,
} from '../type';
import { useTestSourceConnection } from '../useTestSourceConnection';
import { validateGitConnection, validateVaultConnection } from '../validation';

export function ConnectionTest() {
  const { values, setFieldValue } = useFormikContext<FormValues>();

  const livePayload = buildLivePayload(values);

  const debouncedPayload = useDebouncedValue(livePayload);
  const query = useTestSourceConnection(debouncedPayload);

  const settled = isEqual(debouncedPayload, livePayload) && !query.isFetching;
  const connectionOk = settled && query.data?.success === true;

  useEffect(() => {
    if (values.type === 'vault') {
      setFieldValue('vault.connectionOk', connectionOk);
      return;
    }

    setFieldValue('git.connectionOk', connectionOk);
  }, [connectionOk, setFieldValue, values.type]);

  if (!livePayload) {
    return null;
  }

  if (!settled) {
    return (
      <Alert color="info" title="Testing connection...">
        Checking that Portainer can reach the source.
      </Alert>
    );
  }

  if (query.isError) {
    return (
      <Alert color="error" title="Connection failed">
        Unable to test the connection. Please try again.
      </Alert>
    );
  }

  if (query.data?.success) {
    return (
      <Alert color="success" title="Connection successful">
        Portainer reached the source with these details.
      </Alert>
    );
  }

  return (
    <Alert color="error" title="Connection failed">
      {query.data?.error || 'Unable to reach the source.'}
    </Alert>
  );
}

function buildLivePayload(values: FormValues) {
  if (values.type === 'vault') {
    return validateVaultConnection().isValidSync(values.vault)
      ? {
          type: 'vault' as const,
          vault: vaultFormValuesToTestPayload(values.vault),
        }
      : undefined;
  }

  return validateGitConnection().isValidSync(values.git)
    ? {
        type: 'git' as const,
        git: gitFormValuesToTestPayload(values.git),
      }
    : undefined;
}
