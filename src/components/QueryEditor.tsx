import React from 'react';
import { InlineField, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { StripeDataSourceOptions, StripeQuery, QueryType, QUERY_TYPES } from '../types';

type Props = QueryEditorProps<DataSource, StripeQuery, StripeDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTypeChange = (value: SelectableValue<QueryType>) => {
    onChange({ ...query, queryType: value.value! });
    onRunQuery();
  };

  const options = QUERY_TYPES.map((qt) => ({
    label: qt.label,
    value: qt.value,
    description: qt.description,
  }));

  const selected = options.find((o) => o.value === query.queryType) || options[0];

  return (
    <InlineField label="Metric" labelWidth={12} tooltip="Select the Stripe metric to query">
      <Select
        id="query-editor-metric"
        options={options}
        value={selected}
        onChange={onQueryTypeChange}
        width={40}
      />
    </InlineField>
  );
}
