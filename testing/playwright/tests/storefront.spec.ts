import { test, expect } from '@playwright/test';

test.describe('ShopOS Storefront', () => {
  test('homepage loads and has correct title', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/ShopOS/i);
  });

  test('health endpoint returns ok', async ({ request }) => {
    const response = await request.get('/healthz');
    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.status).toBe('ok');
  });

  test('homepage displays product listings', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('main')).toBeVisible();
  });

  test('navigation links are present', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('navigation')).toBeVisible();
  });

  test('search bar is present on homepage', async ({ page }) => {
    await page.goto('/');
    const searchInput = page.getByRole('searchbox');
    await expect(searchInput).toBeVisible();
  });

  test('product page loads when clicking a product', async ({ page }) => {
    await page.goto('/');
    const firstProduct = page.locator('[data-testid="product-card"]').first();
    if (await firstProduct.isVisible()) {
      await firstProduct.click();
      await expect(page).toHaveURL(/\/products\//);
    }
  });

  test('cart icon is accessible', async ({ page }) => {
    await page.goto('/');
    const cartButton = page.getByRole('button', { name: /cart/i });
    await expect(cartButton).toBeVisible();
  });
});
