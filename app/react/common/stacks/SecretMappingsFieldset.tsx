import { useEffect } from 'react';
import { KeyRound, Plus, Trash2 } from 'lucide-react';

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
  const sources = (sourcesQuery.data?.data ?? []).filter(
    (source) => source.type === 'vault'
  );
  const defaultSourceId = sources[0]?.id ?? 0;

  useEffect(() => {
    if (!defaultSourceId || values.every((mapping) => mapping.sourceId)) {
      return;
    }

    onChange(
      values.map((mapping) =>
        mapping.sourceId ? mapping : { ...mapping, sourceId: defaultSourceId }
      )
    );
  }, [defaultSourceId, onChange, values]);

  return (
    <div className="form-group">
      <div className="col-sm-12">
        <div className="mb-3 flex items-center justify-between gap-3">
          <div>
            <div className="control-label !p-0 text-left">Vault secrets</div>
            <p className="mb-0 text-sm text-gray-7 th-highcontrast:text-white th-dark:text-gray-6">
              Choose the Vault paths and keys to pull during deployment.
            </p>
          </div>
          <Button
            type="button"
            color="default"
            icon={Plus}
            onClick={() =>
              onChange([
                ...values,
                { ...emptyMapping, sourceId: defaultSourceId },
              ])
            }
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
              key={index}
              mapping={mapping}
              sources={sources}
              isLoadingSources={sourcesQuery.isLoading}
              index={index}
              error={getRowError(errors, index)}
              providerError={
                !sourcesQuery.isLoading && !defaultSourceId
                  ? 'Vault provider is required'
                  : undefined
              }
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
  isLoadingSources,
  index,
  error,
  providerError,
  onChange,
  onRemove,
}: {
  mapping: StackSecretMapping;
  sources: Source[];
  isLoadingSources: boolean;
  index: number;
  error?: MappingError;
  providerError?: string;
  onChange(value: StackSecretMapping): void;
  onRemove(): void;
}) {
  const selectedSource = sources.find(
    (source) => source.id === mapping.sourceId
  );

  return (
    <div className="rounded border border-solid border-gray-5 bg-gray-1 p-4 th-dark:border-gray-7 th-dark:bg-gray-10">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)_minmax(0,1fr)_auto] md:items-start">
        <FormControl
          inputId={`secret-provider-${index}`}
          label="Vault provider"
          required
          errors={error?.sourceId}
          size="vertical"
          className="mb-0"
        >
          <Select<Source>
            inputId={`secret-provider-${index}`}
            value={selectedSource ?? null}
            options={sources}
            getOptionLabel={(source) => source.name}
            getOptionValue={(source) => String(source.id)}
            onChange={(source) =>
              onChange({ ...mapping, sourceId: source?.id ?? 0 })
            }
            isLoading={isLoadingSources}
            isClearable
            noOptionsMessage={() => 'No Vault providers available'}
            placeholder="Select provider"
            data-cy="secret-mapping-provider-select"
          />
        </FormControl>

        <FormControl
          inputId={`secret-path-${index}`}
          label={
            <span className="inline-flex items-center gap-1">
              <KeyRound className="h-4 w-4" aria-hidden="true" />
              Secret path
            </span>
          }
          required
          errors={error?.path}
          size="vertical"
          className="mb-0"
        >
          <Input
            id={`secret-path-${index}`}
            value={mapping.path}
            placeholder="secret/app"
            onChange={({ target: { value } }) =>
              onChange({ ...mapping, path: value })
            }
            data-cy="secret-mapping-path-input"
          />
        </FormControl>

        <FormControl
          inputId={`secret-key-${index}`}
          label="Secret key"
          required
          errors={error?.key}
          size="vertical"
          className="mb-0"
        >
          <Input
            id={`secret-key-${index}`}
            value={mapping.key}
            placeholder="password"
            onChange={({ target: { value } }) =>
              onChange({ ...mapping, key: value, name: value })
            }
            data-cy="secret-mapping-key-input"
          />
        </FormControl>

        <Button
          type="button"
          color="dangerlight"
          icon={Trash2}
          className="mt-0 md:mt-7"
          onClick={onRemove}
          data-cy="remove-secret-mapping-button"
        />
      </div>
      {(providerError || error?.name) && (
        <p className="small text-danger mb-0 mt-3" role="alert">
          {providerError || error?.name}
        </p>
      )}
    </div>
  );
}
