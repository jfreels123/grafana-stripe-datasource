import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { StripeQuery, StripeDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<StripeQuery, StripeDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<StripeDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<StripeQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: StripeQuery): boolean {
    return !!query.queryType;
  }
}
