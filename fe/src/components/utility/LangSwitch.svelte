<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let lang: string = 'en-US';
	const dispatch = createEventDispatcher();
	const handleLocaleChange = (e: Event) => {
		e.preventDefault();
		dispatch('localeChange', { lang });
	};
</script>

<div class="lang-switch">
	<div class="dropdown-small">
		<select value={lang} on:change={handleLocaleChange}>
			<option value="en-US">English (U.S.)</option>
			<option value="pl">Polski</option>
		</select>
	</div>
</div>

<style>
	.lang-switch {
		position: relative;
		display: flex;
		float: right;
		scale: 95%;
	}
</style>
