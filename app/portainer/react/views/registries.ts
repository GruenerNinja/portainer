import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { ListView } from '@/react/portainer/registries/ListView';
import { ListView as EnvironmentListView } from '@/react/portainer/registries/environments/ListView';
import { RepositoriesView } from '@/react/portainer/registries/repositories/ListView/RepositoriesView';
import { RepositoryView } from '@/react/portainer/registries/repositories/ItemView/RepositoryView';

export const registriesModule = angular
  .module('portainer.app.react.views.registries', [])
  .component(
    'registriesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ListView))), [])
  )
  .component(
    'environmentRegistriesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EnvironmentListView))), [])
  )
  .component(
    'registryRepositoriesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(RepositoriesView))), [])
  )
  .component(
    'registryRepositoryView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(RepositoryView))), [])
  ).name;
