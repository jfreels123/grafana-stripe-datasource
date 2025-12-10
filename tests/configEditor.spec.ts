import { test, expect } from '@grafana/plugin-e2e';
import { StripeDataSourceOptions, StripeSecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByPlaceholder('sk_... or rk_... (secret or restricted key)')).toBeVisible();
});

test('"Save & test" should fail when API key is missing', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource<StripeDataSourceOptions, StripeSecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await expect(configPage.saveAndTest()).not.toBeOK();
});
