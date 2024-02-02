import { getLocaleFromNavigator, init, register } from 'svelte-i18n';
register('en-US', () => import('./data/en_US.json'));
register('pl', () => import('./data/pl.json'));
register('id', () => import('./data/id.json'));
register('de', () => import('./data/de.json'))
register('nl', () => import('./data/nl.json'))
init({
  fallbackLocale: 'en-US',
  initialLocale: getLocaleFromNavigator()
});
