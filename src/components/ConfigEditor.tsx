import React, { ChangeEvent } from 'react';
import { FieldSet, InlineField, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { StripeDataSourceOptions, StripeSecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<StripeDataSourceOptions, StripeSecureJsonData>;

export function ConfigEditor({ options, onOptionsChange }: Props) {
  const { secureJsonFields, secureJsonData } = options;

  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: { apiKey: event.target.value },
    });
  };

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...secureJsonFields, apiKey: false },
      secureJsonData: { ...secureJsonData, apiKey: '' },
    });
  };

  return (
    <FieldSet label="Stripe API">
      <InlineField label="API Key" labelWidth={12} tooltip="Your Stripe secret API key (sk_...)">
        <SecretInput
          required
          id="config-editor-api-key"
          isConfigured={secureJsonFields.apiKey}
          value={secureJsonData?.apiKey || ''}
          placeholder="sk_... or rk_... (secret or restricted key)"
          width={40}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
    </FieldSet>
  );
}
