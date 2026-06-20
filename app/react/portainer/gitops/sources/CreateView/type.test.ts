import { formValuesToCreatePayload, gitFormValuesToTestPayload } from './type';

const baseGit = {
  url: 'https://github.com/org/repo.git',
  tlsSkipVerify: false,
  connectionOk: false,
};

const baseVault = {
  address: '',
  namespace: '',
  kvVersion: 2 as const,
  tlsSkipVerify: false,
  authentication: {
    method: 'token' as const,
    token: '',
  },
  connectionOk: false,
};

function expectGitPayload(
  payload: ReturnType<typeof formValuesToCreatePayload>
) {
  expect(payload.type).toBe('git');
  if (payload.type !== 'git') {
    throw new Error('expected git payload');
  }
  return payload.git;
}

describe('formValuesToCreatePayload', () => {
  it('populates authentication when authEnabled with username and password', () => {
    const payload = expectGitPayload(
      formValuesToCreatePayload({
        name: 'my-source',
        type: 'git',
        git: {
          ...baseGit,
          authentication: {
            authEnabled: true,
            username: 'alice',
            password: 'secret',
          },
        },
        vault: baseVault,
      })
    );

    expect(payload.authentication).toEqual({
      username: 'alice',
      password: 'secret',
    });
  });

  it('omits authentication when authEnabled is false', () => {
    const payload = expectGitPayload(
      formValuesToCreatePayload({
        name: 'my-source',
        type: 'git',
        git: {
          ...baseGit,
          authentication: { authEnabled: false },
        },
        vault: baseVault,
      })
    );

    expect(payload.authentication).toBeUndefined();
  });

  it('omits authentication when authEnabled but username is missing', () => {
    const payload = expectGitPayload(
      formValuesToCreatePayload({
        name: 'my-source',
        type: 'git',
        git: {
          ...baseGit,
          authentication: {
            authEnabled: true,
            password: 'secret',
          },
        },
        vault: baseVault,
      })
    );

    expect(payload.authentication).toBeUndefined();
  });

  it('omits authentication when authEnabled but password is missing', () => {
    const payload = expectGitPayload(
      formValuesToCreatePayload({
        name: 'my-source',
        type: 'git',
        git: {
          ...baseGit,
          authentication: {
            authEnabled: true,
            username: 'alice',
          },
        },
        vault: baseVault,
      })
    );

    expect(payload.authentication).toBeUndefined();
  });

  it('does not include connectionOk in the create payload', () => {
    const payload = expectGitPayload(
      formValuesToCreatePayload({
        name: 'my-source',
        type: 'git',
        git: {
          ...baseGit,
          connectionOk: true,
          authentication: { authEnabled: false },
        },
        vault: baseVault,
      })
    );

    expect(payload).not.toHaveProperty('connectionOk');
  });
});

describe('gitFormValuesToTestPayload', () => {
  it('populates authentication when authEnabled with username and password', () => {
    const payload = gitFormValuesToTestPayload({
      ...baseGit,
      authentication: {
        authEnabled: true,
        username: 'alice',
        password: 'secret',
      },
    });

    expect(payload.authentication).toEqual({
      username: 'alice',
      password: 'secret',
    });
  });

  it('omits authentication when authEnabled is false', () => {
    const payload = gitFormValuesToTestPayload({
      ...baseGit,
      authentication: { authEnabled: false },
    });

    expect(payload.authentication).toBeUndefined();
  });
});
