import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type QueryType =
  | 'mrr' | 'arr' | 'subscribers' | 'customers' | 'balance'
  | 'subscriptions' | 'revenue' | 'invoices' | 'charges' | 'products'
  | 'new_mrr' | 'churned_mrr' | 'net_new_mrr' | 'churn_rate' | 'arpu' | 'trialing' | 'past_due';

export interface StripeQuery extends DataQuery {
  queryType: QueryType;
}

export const DEFAULT_QUERY: Partial<StripeQuery> = {
  queryType: 'mrr',
};

export const QUERY_TYPES: Array<{ label: string; value: QueryType; description: string }> = [
  // Revenue metrics
  { label: 'MRR', value: 'mrr', description: 'Monthly Recurring Revenue' },
  { label: 'ARR', value: 'arr', description: 'Annual Recurring Revenue' },
  { label: 'New MRR', value: 'new_mrr', description: 'MRR from new subscriptions (last 30 days)' },
  { label: 'Churned MRR', value: 'churned_mrr', description: 'MRR lost from cancellations (last 30 days)' },
  { label: 'Net New MRR', value: 'net_new_mrr', description: 'New MRR minus Churned MRR' },
  { label: 'Total Revenue', value: 'revenue', description: 'Total revenue from paid invoices' },
  { label: 'ARPU', value: 'arpu', description: 'Average Revenue Per User' },
  // Subscriber metrics
  { label: 'Active Subscribers', value: 'subscribers', description: 'Count of active subscriptions' },
  { label: 'Churn Rate %', value: 'churn_rate', description: 'Subscriber churn rate (last 30 days)' },
  { label: 'Trialing', value: 'trialing', description: 'Subscriptions in trial' },
  { label: 'Past Due', value: 'past_due', description: 'Subscriptions past due' },
  { label: 'Total Customers', value: 'customers', description: 'Total customer count' },
  // Balance & tables
  { label: 'Available Balance', value: 'balance', description: 'Available balance in USD' },
  { label: 'Subscriptions', value: 'subscriptions', description: 'List of active subscriptions' },
  { label: 'Invoices', value: 'invoices', description: 'List of recent invoices' },
  { label: 'Charges', value: 'charges', description: 'List of recent charges' },
  { label: 'Revenue by Product', value: 'products', description: 'MRR breakdown by product' },
];

export interface StripeDataSourceOptions extends DataSourceJsonData {}

export interface StripeSecureJsonData {
  apiKey?: string;
}
