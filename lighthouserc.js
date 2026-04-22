module.exports = {
  ci: {
    collect: {
      urls: [
        'http://localhost:3000/',
        'http://localhost:3000/products',
        'http://localhost:3000/search',
      ],
      startServerCommand: 'cd src/web/storefront && npm start',
      startServerReadyPattern: 'ready',
      numberOfRuns: 3,
    },
    assert: {
      assertions: {
        'categories:performance':    ['warn',  { minScore: 0.8 }],
        'categories:accessibility':  ['error', { minScore: 0.9 }],
        'categories:best-practices': ['warn',  { minScore: 0.9 }],
        'categories:seo':            ['warn',  { minScore: 0.8 }],
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
}
