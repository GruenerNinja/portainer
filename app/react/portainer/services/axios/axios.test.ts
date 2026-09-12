import { AxiosHeaders, InternalAxiosRequestConfig } from 'axios';

import {
  resetAgentHeaders,
  setPortainerAgentTargetHeader,
} from '@/portainer/services/http-request.helper';

import { agentInterceptor, agentTargetHeader } from './axios';

beforeEach(() => {
  resetAgentHeaders();
});

afterEach(() => {
  resetAgentHeaders();
});

test('agentInterceptor preserves an explicit agent target', () => {
  setPortainerAgentTargetHeader('stale-worker');
  const config = dockerRequest({ [agentTargetHeader]: 'current-worker' });

  const result = agentInterceptor(config);

  expect(result.headers.get(agentTargetHeader)).toBe('current-worker');
});

test('agentInterceptor accepts plain headers from AngularJS adapters', () => {
  setPortainerAgentTargetHeader('stale-worker');
  const config = {
    url: '/endpoints/15/docker/containers/example/json',
    headers: { [agentTargetHeader]: 'current-worker' },
  } as unknown as InternalAxiosRequestConfig;

  const result = agentInterceptor(config);

  expect(result.headers[agentTargetHeader]).toBe('current-worker');
});

test('agentInterceptor uses the legacy queued target when none is explicit', () => {
  setPortainerAgentTargetHeader('legacy-worker');
  const config = dockerRequest();

  const result = agentInterceptor(config);

  expect(result.headers.get(agentTargetHeader)).toBe('legacy-worker');
});

function dockerRequest(headers: Record<string, string> = {}) {
  return {
    url: '/endpoints/15/docker/containers/example/json',
    headers: new AxiosHeaders(headers),
  } as InternalAxiosRequestConfig;
}
