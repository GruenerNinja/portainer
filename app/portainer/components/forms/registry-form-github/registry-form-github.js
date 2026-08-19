class controller {
  constructor($scope) {
    this.$scope = $scope;
    this.toggleOrganisation = this.toggleOrganisation.bind(this);
  }

  $postLink() {
    this.registryFormGithub.registry_name.$validators.used = (modelValue) => !this.nameIsUsed(modelValue);
  }

  toggleOrganisation(newValue) {
    this.$scope.$evalAsync(() => {
      this.model.Github.useOrganisation = newValue;
    });
  }
}

angular.module('portainer.app').component('registryFormGithub', {
  templateUrl: './registry-form-github.html',
  bindings: {
    model: '=',
    formAction: '<',
    formActionLabel: '@',
    actionInProgress: '<',
    nameIsUsed: '<',
  },
  controller,
});
