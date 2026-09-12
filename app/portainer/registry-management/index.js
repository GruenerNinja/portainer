import { AccessHeaders } from '../authorization-guard';

angular.module('portainer.registrymanagement', []).config(config);

/* @ngInject */
function config($stateRegistryProvider) {
  const registries = {
    name: 'portainer.registries',
    url: '/registries',
    views: {
      'content@': {
        component: 'registriesView',
      },
    },
    data: {
      docs: '/admin/registries',
      access: AccessHeaders.Admin,
    },
  };

  const registryCreation = {
    name: 'portainer.registries.new',
    url: '/new',
    views: {
      'content@': {
        component: 'createRegistry',
      },
    },
    data: {
      docs: '/admin/registries/add',
    },
  };

  const registry = {
    name: 'portainer.registries.registry',
    url: '/:id',
    views: {
      'content@': {
        component: 'editRegistry',
      },
    },
    data: {
      docs: '/admin/registries/edit',
    },
  };

  const registryRepositories = {
    name: 'portainer.registries.registry.repositories',
    url: '/repositories?endpointId',
    views: {
      'content@': {
        component: 'registryRepositoriesView',
      },
    },
    data: {
      docs: '/admin/registries/browse',
    },
  };

  const registryRepository = {
    name: 'portainer.registries.registry.repository',
    url: '/repository?repository&endpointId',
    views: {
      'content@': {
        component: 'registryRepositoryView',
      },
    },
    data: {
      docs: '/admin/registries/browse',
    },
  };

  $stateRegistryProvider.register(registries);
  $stateRegistryProvider.register(registry);
  $stateRegistryProvider.register(registryRepositories);
  $stateRegistryProvider.register(registryRepository);
  $stateRegistryProvider.register(registryCreation);
}
