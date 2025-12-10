import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor with metric selector', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(page.getByText('MRR')).toBeVisible();
});
