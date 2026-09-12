import { render, screen } from '@testing-library/react';

import { UserViewModel } from '@/portainer/models/user';
import { withUserProvider } from '@/react/test-utils/withUserProvider';
import { withTestQueryProvider } from '@/react/test-utils/withTestQuery';
import { TeamRole } from '@/react/portainer/users/teams/types';

import { TeamMembersList } from './TeamMembersList';

test('renders correctly', () => {
  const queries = renderComponent();

  expect(queries).toBeTruthy();
});

test('keeps team role changes enabled when membership changes are disabled', () => {
  renderComponent({
    users: [new UserViewModel({ Id: 7, Username: 'synced-user' })],
    roles: { 7: TeamRole.Member },
    membershipChangesDisabled: true,
  });

  expect(screen.getByRole('button', { name: /remove/i })).toBeDisabled();
  expect(screen.getByRole('button', { name: /leader/i })).toBeEnabled();
});

function renderComponent(
  props: Partial<Parameters<typeof TeamMembersList>[0]> = {}
) {
  const user = new UserViewModel({ Username: 'user' });

  const Wrapped = withTestQueryProvider(
    withUserProvider(TeamMembersList, user)
  );

  return render(<Wrapped users={[]} roles={{}} teamId={3} {...props} />);
}

test.todo('when users list is empty, add all users button is disabled');
test.todo('filter displays expected users');
