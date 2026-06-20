import { Plus, Trash2 } from 'lucide-react';

import { StackSecretMapping } from '@/react/common/stacks/types';
import { useSources } from '@/react/portainer/gitops/sources/queries/useSources';
import { Source } from '@/react/portainer/gitops/sources/types';

import { Button } from '@@/buttons';
import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { Select } from '@@/form-components/ReactSelect';

type MappingError = {
  name?: string;
  sourceId?: string;
  path?: string;
  key?: string;
};

interface Props {
  values: StackSecretMapping[];
  onChange(value: StackSecretMapping[]): void;
  errors?: string | Array<string | MappingError>;
}

const emptyMapping: StackSecretMapping = {
  name: '',
  sourceId: 0,
  path: '',
  key: '',
};

export function SecretMappingsFieldset({ values, onChange, errors }: Props) {
  const sourcesQuery = useSources({ type: 'vault' });
  const sources = sourcesQuery.data?.data ?? [];

  return (
    <div className="form-group">
      <div className="col-sm-12">
        <div className="mb-3 flex items-center justify-between gap-3">
          <div>
            <div className="control-label !p-0 text-left">
              Vault secret mappings
            </div>
            <p className="mb-0 text-sm text-gray-7 th-highcontrast:text-white th-dark:text-gray-6">
              Resolve values from Vault into stack environment variables during
              deployment.
            </p>
          </div>
          <Button
            type="button"
            color="default"
            icon={Plus}
            onClick={() => onChange([...values, { ...emptyMapping }])}
            data-cy="add-secret-mapping-button"
          >
            Add mapping
          </Button>
        </div>

        {typeof errors === 'string' && (
          <p className="small text-danger" role="alert">
            {errors}
          </p>
        )}

        <div className="space-y-3">
          {values.map((mapping, index) => (
            <SecretMappingRow
              // Rows do not have a stable database ID.
              key={`${mapping.sourceId}-${mapping.name}-${index}`}
              mapping={mapping}
              sources={sources}
              sourceLoading={sourcesQuery.isLoading}
              error={getRowError(errors, index)}
              onChange={(updated) =>
                onChange(
                  values.map((item, i) => (i === index ? updated : item))
                )
              }
              onRemove={() => onChange(values.filter((_, i) => i !== index))}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

function getRowError(
  errors: Props['errors'],
  index: number
): MappingError | undefined {
  if (!Array.isArray(errors)) {
    return undefined;
  }

  const error = errors[index];
  return typeof error === 'string' ? undefined : error;
}

function SecretMappingRow({
  mapping,
  sources,
  sourceLoading,
  error,
  onChange,
  onRemove,
}: {
  mapping: StackSecretMapping;
  sources: Source[];
  sourceLoading: boolean;
  error?: MappingError;
  onChange(value: StackSecretMapping): void;
  onRemove(): void;
}) {
  const selectedSource = sources.find(
    (source) => source.id === mapping.sourceId
  );

  return (
    <div className="rounded border border-solid border-gray-5 p-3 th-dark:border-gray-7">
      <div className="grid grid-cols-1 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto] lg:items-start">
        <FormControl
          inputId={`secret-name-${mapping.name}`}
          label="Environment variable"
          required
          errors={error?.name}
        >
          <Input
            id={`secret-name-${mapping.name}`}
            value={mapping.name}
            placeholder="DATABASE_PASSWORD"
            onChange={({ target: { value } }) =>
              onChange({ ...mapping, name: value })
            }
            data-cy="secret-mapping-name-input"
          />
        </FormControl>

        <FormControl
          inputId={`secret-source-${mapping.name}`}
          label="Vault source"
          required
          errors={error?.sourceId}
        >
          <Select
            inputId={`secret-source-${mapping.name}`}
            placeholder="Select source"
            value={selectedSource ?? null}
            options={sources}
            getOptionLabel={(source) => source.name}
            getOptionValue={(source) => String(source.id)}
            onChange={(source) =>
              onChange({ ...mapping, sourceId: source?.id ?? 0 })
            }
            isLoading={sourceLoading}
            noOptionsMessage={() => 'No Vault sources available'}
            data-cy="secret-mapping-source-select"
          />
        </FormControl>

        <FormControl
          inputId={`secret-path-${mapping.name}`}
          label="Secret path"
          required
          errors={error?.path}
        >
          <Input
            id={`secret-path-${mapping.name}`}
            value={mapping.path}
            placeholder="secret/app"
            onChange={({ target: { value } }) =>
              onChange({ ...mapping, path: value })
            }
            data-cy="secret-mapping-path-input"
          />
        </FormControl>

        <FormControl
          inputId={`secret-key-${mapping.name}`}
          label="Key"
          required
          errors={error?.key}
        >
          <Input
            id={`secret-key-${mapping.name}`}
            value={mapping.key}
            placeholder="password"
            onChange={({ target: { value } }) =>
              onChange({ ...mapping, key: value })
            }
            data-cy="secret-mapping-key-input"
          />
        </FormControl>

        <Button
          type="button"
          color="dangerlight"
          icon={Trash2}
          className="mt-0 lg:mt-7"
          onClick={onRemove}
          data-cy="remove-secret-mapping-button"
        />
      </div>
    </div>
  );
}
