export const defaultLocale = 'en'
export const locales = ['en', 'es', 'fr', 'de', 'ja', 'ar'] as const
export type Locale = (typeof locales)[number]

export async function getDictionary(locale: Locale) {
  try {
    return (await import(`./dictionaries/${locale}.json`)).default as Record<string, string>
  } catch {
    return (await import('./dictionaries/en.json')).default as Record<string, string>
  }
}
