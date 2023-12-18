<script lang="ts">
	import { onMount } from 'svelte';

	let burgerMenu: HTMLDivElement;
	export let showMobileMenu = false;

	$: if (burgerMenu) {
		burgerMenu.setAttribute('aria-hidden', showMobileMenu ? 'false' : 'true');

		if (showMobileMenu) {
			burgerMenu.style.display = 'block';
		} else {
			burgerMenu.style.display = 'none';
		}
	}

	onMount(() => {
		if (burgerMenu) {
			burgerMenu.setAttribute('aria-hidden', 'true');
			burgerMenu.setAttribute('tabindex', '-1');
		}
	});
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
<div bind:this={burgerMenu}>
	<div on:click|stopPropagation role="dialog">
		<slot name="nick" />
		<hr />
		<slot name="settings" />
		<slot name="logout" />
	</div>
</div>

<style>
	div {
		max-width: 32em;
		border-radius: 0.2em;
		border: none;
		padding: 0.25em;
	}
	div::backdrop {
		background: rgba(0, 0, 0, 0.3);
	}
	div > slot {
		padding: 1em;
	}
</style>
