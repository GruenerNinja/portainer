import { SourceDetail } from '../../queries/useSource';

import { ConnectionDetailsWidget } from './ConnectionDetailsWidget';
import { AuthWidget } from './AuthWidget';
import { AutoUpdateWidget } from './AutoUpdateWidget';
import { SyncStatusWidget } from './SyncStatusWidget';
import { SettingsForm } from './EditForm/SettingsForm';

interface Props {
  source: SourceDetail;
  isEditing: boolean;
  onEditingChange: (isEditing: boolean) => void;
}

export function SettingsTab({ source, isEditing, onEditingChange }: Props) {
  if (isEditing) {
    return (
      <SettingsForm source={source} onCancel={() => onEditingChange(false)} />
    );
  }

  return (
    <>
      <ConnectionDetailsWidget source={source} />
      <AuthWidget
        auth={source?.connection.authentication}
        vault={source?.connection.vault}
      />
      <AutoUpdateWidget autoUpdate={source.autoUpdate} />
      <SyncStatusWidget source={source} />
    </>
  );
}
