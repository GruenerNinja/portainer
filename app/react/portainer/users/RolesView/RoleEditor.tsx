import { FormEvent, useState } from 'react';

import { Button } from '@@/buttons';
import { Checkbox } from '@@/form-components/Checkbox';
import { FormControl } from '@@/form-components/FormControl';
import { Modal } from '@@/modals';

import { RolePayload } from './role.service';
import { RbacRole } from './types';

interface Props {
  role: RbacRole | 'new';
  availableAuthorizations: string[];
  isSaving: boolean;
  onDismiss(): void;
  onSubmit(payload: RolePayload): void;
}

export function RoleEditor({
  role,
  availableAuthorizations,
  isSaving,
  onDismiss,
  onSubmit,
}: Props) {
  const existing = role === 'new' ? undefined : role;
  const [name, setName] = useState(existing?.Name || '');
  const [description, setDescription] = useState(existing?.Description || '');
  const [priority, setPriority] = useState(existing?.Priority || 5);
  const [authorizations, setAuthorizations] = useState(
    existing?.Authorizations || {}
  );

  return (
    <Modal aria-label="Role editor" onDismiss={onDismiss} size="xl">
      <Modal.Header
        title={existing ? `Edit ${existing.Name}` : 'Add custom role'}
      />
      <form onSubmit={handleSubmit}>
        <Modal.Body>
          <p className="text-muted">
            Custom roles can be assigned to users or teams on an environment or
            environment group.
          </p>
          <FormControl label="Name" inputId="role-name" required>
            <input
              id="role-name"
              className="form-control"
              value={name}
              maxLength={128}
              onChange={(event) => setName(event.target.value)}
              data-cy="role-name-input"
              required
            />
          </FormControl>
          <FormControl label="Description" inputId="role-description">
            <textarea
              id="role-description"
              className="form-control"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              data-cy="role-description-input"
            />
          </FormControl>
          <FormControl
            label="Priority"
            inputId="role-priority"
            tooltip="When access policies overlap, the role with the higher priority number wins."
          >
            <input
              id="role-priority"
              type="number"
              min={1}
              max={1000}
              className="form-control"
              value={priority}
              onChange={(event) => setPriority(Number(event.target.value))}
              data-cy="role-priority-input"
              required
            />
          </FormControl>
          <div className="mb-3 flex items-center justify-between">
            <h4 className="m-0">Permissions</h4>
            <div className="flex gap-2">
              <Button
                color="link"
                onClick={() => setAll(true)}
                data-cy="role-select-all"
              >
                Select all
              </Button>
              <Button
                color="link"
                onClick={() => setAll(false)}
                data-cy="role-clear-all"
              >
                Clear all
              </Button>
            </div>
          </div>
          <div className="grid max-h-[45vh] grid-cols-1 gap-2 overflow-y-auto rounded border p-3 md:grid-cols-2">
            {availableAuthorizations.map((authorization) => (
              <Checkbox
                key={authorization}
                id={`role-authorization-${authorization}`}
                label={humanize(authorization)}
                title={authorization}
                checked={Boolean(authorizations[authorization])}
                onChange={(event) =>
                  setAuthorizations((current) => ({
                    ...current,
                    [authorization]: event.target.checked,
                  }))
                }
                bold={false}
                data-cy={`role-authorization-${authorization}`}
              />
            ))}
          </div>
        </Modal.Body>
        <Modal.Footer>
          <Button
            color="default"
            onClick={onDismiss}
            data-cy="role-cancel-button"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={isSaving || !name.trim()}
            data-cy="role-save-button"
          >
            {isSaving ? 'Saving...' : 'Save role'}
          </Button>
        </Modal.Footer>
      </form>
    </Modal>
  );

  function setAll(enabled: boolean) {
    setAuthorizations(
      Object.fromEntries(availableAuthorizations.map((key) => [key, enabled]))
    );
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSubmit({
      Name: name.trim(),
      Description: description.trim(),
      Priority: priority,
      Authorizations: Object.fromEntries(
        Object.entries(authorizations).filter(([, enabled]) => enabled)
      ),
    });
  }
}

function humanize(value: string) {
  return value
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/^Docker /, '')
    .replace(/^Kubernetes /, '');
}
