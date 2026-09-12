import { User } from '@/portainer/users/types';

import { UsersSelector } from '@@/UsersSelector';
import { FormControl } from '@@/form-components/FormControl';
import { Link } from '@@/Link';

interface Props {
  name: string;
  users: User[];
  value: number[];
  onChange(value: number[]): void;
  errors?: string | string[];
  label?: string;
  tooltip?: string;
  inputId?: string;
  dataCy?: string;
}

export function UsersField({
  name,
  users,
  value,
  onChange,
  errors,
  label = 'Authorized users',
  tooltip = 'You can select which user(s) will be able to manage this resource.',
  inputId = 'authorized-users-selector',
  dataCy = 'users-selector',
}: Props) {
  return (
    <FormControl
      label={label}
      tooltip={users.length > 0 ? tooltip : undefined}
      inputId={inputId}
      errors={errors}
    >
      {users.length > 0 ? (
        <UsersSelector
          name={name}
          users={users}
          onChange={onChange}
          value={value}
          inputId={inputId}
          dataCy={dataCy}
        />
      ) : (
        <span className="small text-muted">
          You have not yet created any users. Head over to the{' '}
          <Link to="portainer.users" data-cy="access-control-users-link">
            Users view
          </Link>{' '}
          to manage users.
        </span>
      )}
    </FormControl>
  );
}
