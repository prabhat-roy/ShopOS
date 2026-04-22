import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[type="email"]',    'admin@shopos.io')
  await page.fill('input[type="password"]', 'admin')
  await page.click('button[type="submit"]')
})

test('dashboard renders', async ({ page }) => {
  await expect(page.locator('h1')).toContainText('Dashboard')
})

test('sidebar navigation works', async ({ page }) => {
  await page.locator('a', { hasText: 'Orders' }).click()
  await expect(page.locator('h1')).toContainText('Orders')
})

test('products page loads', async ({ page }) => {
  await page.goto('/products')
  await expect(page.locator('h1')).toContainText('Products')
})
