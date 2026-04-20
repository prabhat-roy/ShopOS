'use strict';

jest.mock('../src/repositories/ContentRepository');

const ContentRepository = require('../src/repositories/ContentRepository');
const { CmsService, NotFoundError, ValidationError } = require('../src/services/CmsService');

let service;

beforeEach(() => {
  service = new CmsService();
  jest.clearAllMocks();
});

// ── Test 1: create content ──────────────────────────────────────────────────
test('createContent — creates item with rendered HTML', async () => {
  const data = { title: 'Hello World', type: 'blog_post', body: '# Heading\nText' };
  const saved = { _id: 'id1', ...data, slug: 'hello-world-abc12345', htmlContent: '<h1>Heading</h1>\n<p>Text</p>\n', status: 'draft' };
  ContentRepository.create.mockResolvedValue(saved);

  const result = await service.createContent(data);

  expect(ContentRepository.create).toHaveBeenCalledWith(
    expect.objectContaining({ title: 'Hello World', type: 'blog_post' })
  );
  expect(result.htmlContent).toContain('<h1>');
});

// ── Test 2: slug auto-generation ────────────────────────────────────────────
test('createContent — auto-generates slug from title when not provided', async () => {
  const data = { title: 'My Great Article!', type: 'page', body: '' };
  ContentRepository.create.mockImplementation(async (d) => ({ _id: 'id2', ...d, status: 'draft' }));

  const result = await service.createContent(data);

  expect(result.slug).toMatch(/^my-great-article-/);
});

// ── Test 3: custom slug is used ─────────────────────────────────────────────
test('createContent — uses provided slug', async () => {
  const data = { title: 'Custom Slug Page', type: 'page', body: '', slug: 'custom-slug' };
  ContentRepository.create.mockImplementation(async (d) => ({ _id: 'id3', ...d, status: 'draft' }));

  const result = await service.createContent(data);

  expect(result.slug).toBe('custom-slug');
});

// ── Test 4: validation — missing title ──────────────────────────────────────
test('createContent — throws ValidationError when title is missing', async () => {
  await expect(service.createContent({ type: 'page' })).rejects.toThrow(ValidationError);
});

// ── Test 5: get by slug ──────────────────────────────────────────────────────
test('getBySlug — returns item for matching slug and locale', async () => {
  const item = { _id: 'id4', slug: 'my-page', locale: 'en', status: 'published' };
  ContentRepository.findBySlug.mockResolvedValue(item);

  const result = await service.getBySlug('my-page', 'en');

  expect(result.slug).toBe('my-page');
  expect(ContentRepository.findBySlug).toHaveBeenCalledWith('my-page', 'en');
});

// ── Test 6: get by slug — not found ─────────────────────────────────────────
test('getBySlug — throws NotFoundError when slug does not exist', async () => {
  ContentRepository.findBySlug.mockResolvedValue(null);

  await expect(service.getBySlug('no-such-slug', 'en')).rejects.toThrow(NotFoundError);
});

// ── Test 7: get by id ────────────────────────────────────────────────────────
test('getById — returns item for matching id', async () => {
  const item = { _id: 'id5', title: 'Test', status: 'draft' };
  ContentRepository.findById.mockResolvedValue(item);

  const result = await service.getById('id5');

  expect(result._id).toBe('id5');
});

// ── Test 8: list with filters ────────────────────────────────────────────────
test('listContent — passes filters to repository', async () => {
  const mockResult = { items: [], total: 0, limit: 20, offset: 0 };
  ContentRepository.list.mockResolvedValue(mockResult);

  await service.listContent({ type: 'blog_post', status: 'published', locale: 'en' });

  expect(ContentRepository.list).toHaveBeenCalledWith(
    expect.objectContaining({ type: 'blog_post', status: 'published', locale: 'en' })
  );
});

// ── Test 9: update re-renders HTML ───────────────────────────────────────────
test('updateContent — re-renders HTML when body is updated', async () => {
  const existing = { _id: 'id6', title: 'Old', body: 'old body', status: 'draft' };
  ContentRepository.findById.mockResolvedValue(existing);
  ContentRepository.update.mockImplementation(async (_id, data) => ({ ...existing, ...data }));

  const result = await service.updateContent('id6', { body: '## New Heading' });

  expect(result.htmlContent).toContain('<h2>');
  expect(ContentRepository.update).toHaveBeenCalledWith(
    'id6',
    expect.objectContaining({ htmlContent: expect.stringContaining('<h2>') })
  );
});

// ── Test 10: publish ─────────────────────────────────────────────────────────
test('publishContent — transitions status to published', async () => {
  const existing = { _id: 'id7', status: 'draft' };
  ContentRepository.findById.mockResolvedValue(existing);
  ContentRepository.publish.mockResolvedValue({ ...existing, status: 'published', publishedAt: new Date() });

  await service.publishContent('id7');

  expect(ContentRepository.publish).toHaveBeenCalledWith('id7');
});

// ── Test 11: archive ─────────────────────────────────────────────────────────
test('archiveContent — transitions status to archived', async () => {
  const existing = { _id: 'id8', status: 'published' };
  ContentRepository.findById.mockResolvedValue(existing);
  ContentRepository.archive.mockResolvedValue({ ...existing, status: 'archived' });

  await service.archiveContent('id8');

  expect(ContentRepository.archive).toHaveBeenCalledWith('id8');
});

// ── Test 12: delete draft ────────────────────────────────────────────────────
test('deleteContent — deletes a draft item', async () => {
  const existing = { _id: 'id9', status: 'draft' };
  ContentRepository.findById.mockResolvedValue(existing);
  ContentRepository.delete.mockResolvedValue(existing);

  await service.deleteContent('id9');

  expect(ContentRepository.delete).toHaveBeenCalledWith('id9');
});

// ── Test 13: delete published — fails ────────────────────────────────────────
test('deleteContent — throws ValidationError when trying to delete published content', async () => {
  const existing = { _id: 'id10', status: 'published' };
  ContentRepository.findById.mockResolvedValue(existing);

  await expect(service.deleteContent('id10')).rejects.toThrow(ValidationError);
  expect(ContentRepository.delete).not.toHaveBeenCalled();
});

// ── Test 14: search ──────────────────────────────────────────────────────────
test('searchContent — returns matched results for query', async () => {
  const mockResults = [
    { _id: 'id11', title: 'Spring Boot Guide', score: 1.5 },
    { _id: 'id12', title: 'Spring Security', score: 1.2 },
  ];
  ContentRepository.search.mockResolvedValue(mockResults);

  const results = await service.searchContent('spring', 'en');

  expect(results).toHaveLength(2);
  expect(ContentRepository.search).toHaveBeenCalledWith('spring', 'en');
});
