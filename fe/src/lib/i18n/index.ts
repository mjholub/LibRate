import { getLocaleFromNavigator, init, register } from 'svelte-i18n';
register('en-US', () => import('./data/en_US.json'));
register('pl', () => import('./data/pl.json'));
init({
  fallbackLocale: 'en-US',
  initialLocale: getLocaleFromNavigator()
});
