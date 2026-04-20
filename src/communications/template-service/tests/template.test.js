'use strict';

// Mock mongoose before any module that imports it is loaded
jest.mock('mongoose', () => {
  const actualMongoose = jest.requireActual('mongoose');

  // We'll use an in-memory store to simulate MongoDB operations
  const store = new Map();
  let idCounter = 1;

  function makeId() {
    return String(idCounter++).padStart(24, '0');
  }

  class MockModel {
    constructor(data) {
      Object.assign(this, data);
      if (!this._id) {
        this._id = makeId();
      }
      if (!this.active) this.active = true;
      if (!this.version) this.version = 1;
      if (!this.variables) this.variables = [];
      if (!this.tags) this.tags = [];
      this.createdAt = new Date();
      this.updatedAt = new Date();
    }

    async save() {
      // Check uniqueness on `name`
      for (const [, doc] of store) {
        if (doc.name === this.name && doc._id !== this._id) {
          const err = new Error('Duplicate key error');
          err.code = 11000;
          throw err;
        }
      }
      store.set(this._id, { ...this });
      return { ...this };
    }

    static async findById(id) {
      const doc = store.get(String(id));
      return doc ? { ...doc } : null;
    }

    static async findOne(query) {
      for (const [, doc] of store) {
        let match = true;
        for (const [k, v] of Object.entries(query)) {
          if (doc[k] !== v) { match = false; break; }
        }
        if (match) return { ...doc };
      }
      return null;
    }

    static find(query) {
      const results = [];
      for (const [, doc] of store) {
        let match = true;
        for (const [k, v] of Object.entries(query)) {
          if (v && typeof v === 'object' && v.$in) {
            if (!doc[k] || !doc[k].some((t) => v.$in.includes(t))) { match = false; break; }
          } else if (doc[k] !== v) {
            match = false; break;
          }
        }
        if (match) results.push({ ...doc });
      }
      return {
        sort: () => ({ lean: () => Promise.resolve(results) }),
        lean: () => Promise.resolve(results),
      };
    }

    static async findByIdAndUpdate(id, update, opts) {
      const doc = store.get(String(id));
      if (!doc) return null;
      const $set = update.$set || {};
      const updated = { ...doc, ...$set, updatedAt: new Date() };
      store.set(String(id), updated);
      return opts && opts.new ? { ...updated } : { ...doc };
    }

    static lean() { return this; }
  }

  MockModel._store = store;
  MockModel._reset = () => { store.clear(); idCounter = 1; };

  const mongoose = {
    connect: jest.fn().mockResolvedValue(undefined),
    connection: { close: jest.fn().mockResolvedValue(undefined) },
    Schema: actualMongoose.Schema,
    model: jest.fn().mockReturnValue(MockModel),
    __MockModel: MockModel,
  };

  return mongoose;
});

// Clear the store before each test
const mongoose = require('mongoose');
const MockModel = mongoose.__MockModel;

beforeEach(() => {
  MockModel._reset();
});

const repo = require('../src/repositories/TemplateRepository');
const service = require('../src/services/TemplateService');

// ─── Helper ───────────────────────────────────────────────────────────────────

function emailTemplateData(overrides = {}) {
  return {
    name: 'welcome-email',
    channel: 'email',
    subject: 'Welcome, {{name}}!',
    body: 'Hello {{name}}, your account {{accountId}} is ready.',
    variables: [
      { name: 'name', required: true, defaultValue: null },
      { name: 'accountId', required: false, defaultValue: 'N/A' },
    ],
    tags: ['welcome', 'onboarding'],
    ...overrides,
  };
}

// ─── 1. Create template ───────────────────────────────────────────────────────

test('creates a new template and returns it', async () => {
  const data = emailTemplateData();
  const template = await service.createTemplate(data);

  expect(template).toBeDefined();
  expect(template.name).toBe('welcome-email');
  expect(template.channel).toBe('email');
  expect(template.version).toBe(1);
  expect(template.active).toBe(true);
});

// ─── 2. Get template by id ────────────────────────────────────────────────────

test('retrieves a template by id', async () => {
  const created = await service.createTemplate(emailTemplateData());
  const found = await service.getTemplate(String(created._id));

  expect(found).toBeDefined();
  expect(found._id).toBe(created._id);
  expect(found.name).toBe('welcome-email');
});

// ─── 3. Get template by name ──────────────────────────────────────────────────

test('retrieves an active template by name', async () => {
  await service.createTemplate(emailTemplateData());
  const found = await service.getByName('welcome-email');

  expect(found.name).toBe('welcome-email');
  expect(found.active).toBe(true);
});

// ─── 4. List templates by channel ────────────────────────────────────────────

test('lists templates filtered by channel', async () => {
  await service.createTemplate(emailTemplateData({ name: 'tmpl-email-1' }));
  await service.createTemplate(emailTemplateData({ name: 'tmpl-email-2', channel: 'email' }));
  await service.createTemplate(emailTemplateData({ name: 'tmpl-sms-1', channel: 'sms', subject: null }));

  const emailTemplates = await service.listTemplates({ channel: 'email' });
  expect(emailTemplates.length).toBe(2);
  emailTemplates.forEach((t) => expect(t.channel).toBe('email'));
});

// ─── 5. Update bumps version ──────────────────────────────────────────────────

test('updating a template increments the version', async () => {
  const created = await service.createTemplate(emailTemplateData());
  expect(created.version).toBe(1);

  const updated = await service.updateTemplate(String(created._id), { subject: 'New Subject' });
  expect(updated.version).toBe(2);
  expect(updated.subject).toBe('New Subject');
});

// ─── 6. Archive template ──────────────────────────────────────────────────────

test('archiving a template sets active=false', async () => {
  const created = await service.createTemplate(emailTemplateData());
  const archived = await service.archiveTemplate(String(created._id));

  expect(archived.active).toBe(false);
});

// ─── 7. Render with all variables ─────────────────────────────────────────────

test('renders a template with all variables provided', async () => {
  await service.createTemplate(emailTemplateData());

  const result = await service.renderTemplate('welcome-email', {
    name: 'Alice',
    accountId: 'ACC-001',
  });

  expect(result.subject).toBe('Welcome, Alice!');
  expect(result.body).toBe('Hello Alice, your account ACC-001 is ready.');
  expect(result.channel).toBe('email');
});

// ─── 8. Render uses default values ────────────────────────────────────────────

test('renders a template using defaultValue for missing optional variable', async () => {
  await service.createTemplate(emailTemplateData());

  const result = await service.renderTemplate('welcome-email', { name: 'Bob' });

  expect(result.subject).toBe('Welcome, Bob!');
  expect(result.body).toBe('Hello Bob, your account N/A is ready.');
});

// ─── 9. Render throws for missing required variable ───────────────────────────

test('renderTemplate throws MISSING_REQUIRED_VARIABLE when required var is absent', async () => {
  await service.createTemplate(emailTemplateData());

  await expect(
    service.renderTemplate('welcome-email', { accountId: 'ACC-999' }),
  ).rejects.toMatchObject({ code: 'MISSING_REQUIRED_VARIABLE', variable: 'name' });
});

// ─── 10. Render throws NOT_FOUND for unknown template ─────────────────────────

test('renderTemplate throws NOT_FOUND for non-existent template name', async () => {
  await expect(service.renderTemplate('does-not-exist', {})).rejects.toMatchObject({
    code: 'NOT_FOUND',
  });
});

// ─── 11. getTemplate throws NOT_FOUND for unknown id ─────────────────────────

test('getTemplate throws NOT_FOUND for unknown id', async () => {
  await expect(service.getTemplate('000000000000000000000099')).rejects.toMatchObject({
    code: 'NOT_FOUND',
  });
});

// ─── 12. getByName throws NOT_FOUND for unknown name ─────────────────────────

test('getByName throws NOT_FOUND for unknown template name', async () => {
  await expect(service.getByName('ghost-template')).rejects.toMatchObject({
    code: 'NOT_FOUND',
  });
});

// ─── 13. repo.render substitutes all {{var}} placeholders ─────────────────────

test('repo.render correctly substitutes all placeholders', () => {
  const tmpl = {
    subject: 'Order {{orderId}} confirmed',
    body: 'Hi {{customer}}, order {{orderId}} ships on {{date}}.',
    variables: [],
  };

  const result = repo.render(tmpl, { orderId: 'ORD-123', customer: 'Charlie', date: '2026-05-01' });

  expect(result.subject).toBe('Order ORD-123 confirmed');
  expect(result.body).toBe('Hi Charlie, order ORD-123 ships on 2026-05-01.');
});

// ─── 14. repo.render leaves unknown placeholders as empty string ──────────────

test('repo.render replaces unknown placeholders with empty string', () => {
  const tmpl = {
    subject: null,
    body: 'Hello {{firstName}} {{lastName}}',
    variables: [],
  };

  const result = repo.render(tmpl, { firstName: 'Dave' });

  expect(result.body).toBe('Hello Dave ');
});
