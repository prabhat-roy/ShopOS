// Shared test data for k6 load tests

export const PRODUCT_IDS = [
  'prod-001', 'prod-002', 'prod-003', 'prod-004', 'prod-005',
  'prod-010', 'prod-020', 'prod-030', 'prod-040', 'prod-050',
  'prod-100', 'prod-200', 'prod-300', 'prod-400', 'prod-500',
];

export const CATEGORY_IDS = [
  'cat-electronics', 'cat-clothing', 'cat-books',
  'cat-home', 'cat-sports', 'cat-beauty',
];

export const SEARCH_QUERIES = [
  'laptop', 'smartphone', 'headphones', 'keyboard', 'monitor',
  'shirt', 'shoes', 'jeans', 'jacket', 'dress',
  'python book', 'coffee maker', 'yoga mat', 'running shoes',
];

export const TEST_USERS = [
  { email: 'user1@test.shopos.local', password: 'Password1!' },
  { email: 'user2@test.shopos.local', password: 'Password2!' },
  { email: 'user3@test.shopos.local', password: 'Password3!' },
  { email: 'user4@test.shopos.local', password: 'Password4!' },
  { email: 'user5@test.shopos.local', password: 'Password5!' },
];

export const PAYMENT_METHODS = [
  { type: 'card', number: '4111111111111111', expiry: '12/28', cvv: '123' },
  { type: 'card', number: '5500005555555559', expiry: '11/27', cvv: '456' },
];

export const SHIPPING_ADDRESSES = [
  {
    line1: '123 Test Street', city: 'New York',
    state: 'NY', zip: '10001', country: 'US',
  },
  {
    line1: '456 Load Ave', city: 'San Francisco',
    state: 'CA', zip: '94105', country: 'US',
  },
];
