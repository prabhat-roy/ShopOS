export interface ApiKey    { id: string; name: string; prefix: string; createdAt: string; lastUsed: string | null; scopes: string[] }
export interface Webhook   { id: string; url: string; events: string[]; status: string; createdAt: string; successRate: number }
export interface ApiUsage  { endpoint: string; requests: number; errors: number; p50ms: number; p99ms: number }
export interface AppStats  { totalRequests: number; errorRate: number; activeKeys: number; activeWebhooks: number }
