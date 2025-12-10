import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type QueryType = 'mrr' | 'arr' | 'subscribers' | 'customers' | 'balance' | 'subscriptions';

export interface StripeQuery extends DataQuery {
  queryType: QueryType;
}

export const DEFAULT_QUERY: Partial<StripeQuery> = {
  queryType: 'mrr',
};

export const QUERY_TYPES: Array<{ label: string; value: QueryType; description: string }> = [
  { label: 'MRR', value: 'mrr', description: 'Monthly Recurring Revenue' },
  { label: 'ARR', value: 'arr', description: 'Annual Recurring Revenue' },
  { label: 'Active Subscribers', value: 'subscribers', description: 'Count of active subscriptions' },
  { label: 'Total Customers', value: 'customers', description: 'Total customer count' },
  { label: 'Available Balance', value: 'balance', description: 'Available balance in USD' },
  { label: 'Subscriptions', value: 'subscriptions', description: 'List of active subscriptions' },
];

export interface StripeDataSourceOptions extends DataSourceJsonData {}

export interface StripeSecureJsonData {
  apiKey?: string;
}
