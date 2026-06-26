import { useFormikContext } from 'formik';

import { GitForm } from '@/react/portainer/gitops/GitForm';
import { baseStackWebhookUrl } from '@/portainer/helpers/webhookHelper';
import { SecretMappingsFieldset } from '@/react/common/stacks/SecretMappingsFieldset';

import { FormValues } from '../types';

import { StackRelativePathFieldset } from './StackRelativePathFieldset';

interface Props {
  isDockerStandalone?: boolean;
  webhookId: string;
}

export function GitSection({ webhookId, isDockerStandalone = false }: Props) {
  const { values, errors, setFieldValue, setValues } =
    useFormikContext<FormValues>();

  return (
    <>
      <GitForm
        value={values.git}
        onChange={(gitValues) =>
          setValues((values) => ({
            ...values,
            git: {
              ...values.git,
              ...gitValues,
            },
          }))
        }
        environmentType="DOCKER"
        deployMethod="compose"
        isDockerStandalone={isDockerStandalone}
        isAdditionalFilesFieldVisible
        isForcePullVisible
        errors={errors.git}
        baseWebhookUrl={baseStackWebhookUrl()}
        webhookId={webhookId}
      />
      <StackRelativePathFieldset isDockerStandalone={isDockerStandalone} />
      <SecretMappingsFieldset
        values={values.secretMappings}
        onChange={(value) => setFieldValue('secretMappings', value)}
        errors={
          Array.isArray(errors.secretMappings) ||
          typeof errors.secretMappings === 'string'
            ? errors.secretMappings
            : undefined
        }
      />
    </>
  );
}
