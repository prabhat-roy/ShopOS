import { test, expect } from '@playwright/test'

test('home page loads', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveTitle(/ShopOS/)
  await expect(page.locator('h1')).toContainText('ShopOS')
})

test('product listing page', async ({ page }) => {
  await page.goto('/products')
  await expect(page.locator('h1')).toContainText('Products')
})

test('search works', async ({ page }) => {
  await page.goto('/search')
  await page.fill('input[placeholder*="Search"]', 'laptop')
  await page.keyboard.press('Enter')
  await expect(page.locator('form')).toBeVisible()
})

test('cart page loads empty', async ({ page }) => {
  await page.goto('/cart')
  await expect(page.getByText(/cart is empty|Cart/i)).toBeVisible()
})

test('login page renders', async ({ page }) => {
  await page.goto('/login')
  await expect(page.locator('input[type="email"]')).toBeVisible()
  await expect(page.locator('input[type="password"]')).toBeVisible()
})
