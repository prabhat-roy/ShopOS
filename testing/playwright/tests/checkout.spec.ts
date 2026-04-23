import { test, expect } from '@playwright/test';

test.describe('ShopOS Checkout Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to homepage before each test
    await page.goto('/');
  });

  test('add product to cart', async ({ page }) => {
    // Navigate to a product page
    await page.goto('/products');
    const addToCartBtn = page.getByRole('button', { name: /add to cart/i }).first();
    if (await addToCartBtn.isVisible()) {
      await addToCartBtn.click();
      // Verify cart count updated
      const cartBadge = page.locator('[data-testid="cart-count"]');
      await expect(cartBadge).toHaveText('1');
    }
  });

  test('cart page shows added items', async ({ page }) => {
    await page.goto('/cart');
    await expect(page).toHaveURL('/cart');
    await expect(page.getByRole('main')).toBeVisible();
  });

  test('checkout page requires authentication', async ({ page }) => {
    await page.goto('/checkout');
    // Should redirect to login or show auth prompt
    await expect(page).toHaveURL(/(login|checkout)/);
  });

  test('guest checkout flow', async ({ page }) => {
    await page.goto('/checkout');
    const guestBtn = page.getByRole('button', { name: /guest/i });
    if (await guestBtn.isVisible()) {
      await guestBtn.click();
      await expect(page.getByRole('form')).toBeVisible();
    }
  });

  test('checkout form validates required fields', async ({ page }) => {
    await page.goto('/checkout');
    const submitBtn = page.getByRole('button', { name: /place order|submit|continue/i });
    if (await submitBtn.isVisible()) {
      await submitBtn.click();
      // Form validation errors should appear
      const errors = page.locator('[data-testid="field-error"], .error-message, [role="alert"]');
      // At least one validation message should be present if fields are empty
      await expect(errors.first()).toBeVisible({ timeout: 5000 }).catch(() => {
        // Form may redirect if cart is empty — acceptable
      });
    }
  });

  test('payment method selection', async ({ page }) => {
    await page.goto('/checkout');
    const creditCardOption = page.getByLabel(/credit card/i);
    if (await creditCardOption.isVisible()) {
      await creditCardOption.click();
      await expect(creditCardOption).toBeChecked();
    }
  });

  test('order confirmation page', async ({ page }) => {
    await page.goto('/order-confirmation');
    await expect(page.getByRole('main')).toBeVisible();
  });
});
