import { test, expect } from '@playwright/test'

test('login page renders', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByText('Developer Login')).toBeVisible()
  await expect(page.locator('input[type="email"]')).toBeVisible()
})

test.describe('authenticated', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="email"]',    'dev@shopos.io')
    await page.fill('input[type="password"]', 'password')
    await page.click('button[type="submit"]')
  })

  test('dashboard loads', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Dashboard')
  })

  test('api keys page loads', async ({ page }) => {
    await page.goto('/api-keys')
    await expect(page.locator('h1')).toContainText('API Keys')
  })

  test('sandbox page loads', async ({ page }) => {
    await page.goto('/sandbox')
    await expect(page.locator('h1')).toContainText('Sandbox')
  })
})
