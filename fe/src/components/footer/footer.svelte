<script lang="ts">
	import { locale, _ } from 'svelte-i18n';

	const currentYear = new Date().getFullYear();
	$: tos_url = `/tos_${$locale}.html`;
	$: privacy_url = `/privacy_${$locale}.html`;
</script>

<footer>
	<p>
		Copyright 2023-{currentYear}
		<a href="https://codeberg.org/mjh/LibRate/src/branch/main/CONTRIBUTORS.md"
			>LibRate contributors</a
		>
	</p>
	<p>
		Licensed under the <link rel="license" href="https://gnu.org/licenses/agpl-3.0.html" />AGPL-3.0
	</p>
	<p><a href={tos_url}>{$_('tos')}</a> | <a href={privacy_url}>{$_('privacy_policy')}</a></p>
</footer>

<style>
	footer {
		text-align: center;
		padding: 0.5rem;
		font-size: 0.8rem;
		color: #fff;
		background-color: #333;
	}
	footer p {
		margin-bottom: 0.4rem;
	}
</style>
