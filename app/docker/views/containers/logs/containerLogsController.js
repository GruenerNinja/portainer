import moment from 'moment';

import { openDockerLogsStream } from '@/docker/helpers/logHelper';

angular.module('portainer.docker').controller('ContainerLogsController', [
  '$scope',
  '$transition$',
  'ContainerService',
  'Notifications',
  'HttpRequestHelper',
  'endpoint',
  function ($scope, $transition$, ContainerService, Notifications, HttpRequestHelper, endpoint) {
    let logStream;
    let collectionEnabled = true;
    let settingsWatchReady = false;

    $scope.logs = [];
    $scope.state = {
      lineCount: 100,
      sinceTimestamp: '',
      displayTimestamps: false,
    };

    $scope.changeLogCollection = function (logCollectionStatus) {
      collectionEnabled = logCollectionStatus;
      if (collectionEnabled) {
        startLogStream();
      } else {
        stopLogStream();
      }
    };

    const stopSettingsWatch = $scope.$watchGroup(['state.displayTimestamps', 'state.sinceTimestamp', 'state.lineCount'], function () {
      if (!settingsWatchReady) {
        settingsWatchReady = true;
        return;
      }
      if (collectionEnabled && $scope.container) {
        startLogStream();
      }
    });

    $scope.$on('$destroy', function () {
      stopSettingsWatch();
      stopLogStream();
    });

    function stopLogStream() {
      logStream?.close();
      logStream = undefined;
    }

    function startLogStream() {
      stopLogStream();
      const since = moment($scope.state.sinceTimestamp).unix();
      logStream = openDockerLogsStream({
        environmentId: endpoint.Id,
        resource: 'containers',
        resourceId: $transition$.params().id,
        nodeName: $transition$.params().nodeName,
        timestamps: $scope.state.displayTimestamps,
        since: Number.isFinite(since) ? since : 0,
        tail: $scope.state.lineCount,
        multiplexed: !$scope.container.Config.Tty,
        onLogs(logs) {
          $scope.$evalAsync(() => {
            $scope.logs = logs;
          });
        },
        onError(error) {
          $scope.$evalAsync(() => {
            stopLogStream();
            Notifications.error('Failure', error, 'Unable to stream container logs');
          });
        },
      });
    }

    function initView() {
      HttpRequestHelper.setPortainerAgentTargetHeader($transition$.params().nodeName);
      ContainerService.container(endpoint.Id, $transition$.params().id)
        .then(function success(container) {
          $scope.container = container;
          $scope.logsEnabled = container.HostConfig?.LogConfig?.Type && container.HostConfig.LogConfig.Type !== 'none';

          if ($scope.logsEnabled) {
            startLogStream();
          }
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve container information');
        });
    }

    initView();
  },
]);
