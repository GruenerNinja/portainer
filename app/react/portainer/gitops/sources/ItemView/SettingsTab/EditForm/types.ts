import { boolean as yupBoolean, number, object, string } from 'yup';

import { intervalValidation } from '../../../components/IntervalField';

export interface SettingsFormValues {
  type: 'git' | 'vault';
  name: string;
  url: string;
  tlsSkipVerify: boolean;
  authEnabled: boolean;
  username: string;
  password: string;
  namespace: string;
  kvVersion: number;
  token: string;
  pollingEnabled: boolean;
  interval: string;
}

export const validationSchema = object({
  type: string().oneOf(['git', 'vault']).required(),
  name: string().required('Name is required'),
  url: string().required('URL is required'),
  tlsSkipVerify: yupBoolean().defined(),
  authEnabled: yupBoolean().defined(),
  username: string().when('authEnabled', {
    is: true,
    then: (schema) => schema.required('Username is required'),
    otherwise: (schema) => schema.optional(),
  }),
  password: string().optional(),
  namespace: string().optional(),
  kvVersion: number().when('type', {
    is: 'vault',
    then: (schema) => schema.required('KV version is required'),
  }),
  token: string().optional(),
  pollingEnabled: yupBoolean().defined(),
  interval: string()
    .default('')
    .when('pollingEnabled', {
      is: true,
      then: () => intervalValidation(),
    }),
});
