import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { StripeQuery, StripeDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, StripeQuery, StripeDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
