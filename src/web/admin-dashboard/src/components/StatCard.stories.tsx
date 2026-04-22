import type { Meta, StoryObj } from '@storybook/react'
import StatCard from './StatCard'

const meta: Meta<typeof StatCard> = {
  title: 'Components/StatCard',
  component: StatCard,
  tags: ['autodocs'],
}
export default meta
type Story = StoryObj<typeof StatCard>

export const Revenue: Story = { args: { label: 'Total Revenue', value: 125430.50, prefix: '$', change: 12.3 } }
export const Orders:  Story = { args: { label: 'Total Orders',  value: 2847,     change: -3.1 } }
export const NoChange: Story = { args: { label: 'Total Products', value: 450 } }
