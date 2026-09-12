import moment from 'moment';

import { openDockerLogsStream } from '@/docker/helpers/logHelper';

angular.module('portainer.docker').controller('TaskLogsController', [
  '$scope',
  '$transition$',
  'TaskService',
  'ServiceService',
  'Notifications',
  'endpoint',
  function ($scope, $transition$, TaskService, ServiceService, Notifications, endpoint) {
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
      if (collectionEnabled && $scope.service) {
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
        resource: 'tasks',
        resourceId: $transition$.params().id,
        timestamps: $scope.state.displayTimestamps,
        since: Number.isFinite(since) ? since : 0,
        tail: $scope.state.lineCount,
        multiplexed: true,
        onLogs(logs) {
          $scope.$evalAsync(() => {
            $scope.logs = logs;
          });
        },
        onError(error) {
          $scope.$evalAsync(() => {
            stopLogStream();
            Notifications.error('Failure', error, 'Unable to stream task logs');
          });
        },
      });
    }

    function initView() {
      TaskService.task($transition$.params().id)
        .then(function success(task) {
          $scope.task = task;
          return ServiceService.service(task.ServiceId);
        })
        .then(function success(service) {
          $scope.service = service;
          startLogStream();
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve task details');
        });
    }

    initView();
  },
]);
