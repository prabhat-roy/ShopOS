import type { Meta, StoryObj } from '@storybook/react'
import ProductCard from './ProductCard'

const meta: Meta<typeof ProductCard> = {
  title: 'Components/ProductCard',
  component: ProductCard,
  tags: ['autodocs'],
}
export default meta

type Story = StoryObj<typeof ProductCard>

export const Default: Story = {
  args: {
    product: {
      id: 'prod_1',
      name: 'Premium Wireless Headphones',
      slug: 'premium-wireless-headphones',
      price: 149.99,
      currency: 'USD',
      imageUrl: 'https://placehold.co/400x400',
      category: 'Electronics',
      rating: 4.5,
      reviewCount: 128,
      inStock: true,
    },
  },
}

export const OutOfStock: Story = {
  args: {
    product: { ...Default.args!.product!, inStock: false },
  },
}
