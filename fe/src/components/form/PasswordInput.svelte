<script lang="ts">
	import { EyeIcon, EyeOffIcon } from 'svelte-feather-icons';

	export let value: string;
	export let id: string;
	export let onInput: (password: string) => Promise<void>;
	export let showPassword: boolean;
	export let toggleObfuscation: () => void;
</script>

<label for={id}>Password:</label>
<div class="password-container">
	<input
		{id}
		class={!showPassword ? '' : 'hidden'}
		bind:value
		type="password"
		autocomplete="new-password"
		required
		on:input={() => {
			onInput(value);
		}}
	/>
	<input
		id="{id}Text"
		class={showPassword ? '' : 'hidden'}
		bind:value
		type="text"
		autocomplete="new-password"
		required
		aria-label="Password"
	/>
	<button
		class="toggle-btn"
		type="button"
		on:click|preventDefault={toggleObfuscation}
		aria-label="Toggle password visibility"
	>
		{#if showPassword}
			<EyeIcon />
		{:else}
			<EyeOffIcon />
		{/if}
	</button>
</div>

<style>
	:root {
		--input-border-color: #ccc;
		--input-border-color-focus: #aaa;
		--input-border-color-error: #f00;
		--input-background-color: #fff;
		--input-background-color-focus: #fff;
		--input-text: #000;
		--border-radius: 2px;
		--pwd-container-display: inline flow-root list-item;
	}

	.password-container {
		position: relative;
		overflow: hidden;
		display: var(--pwd-container-display);
		border: 1px solid var(--input-border-color);
		border-radius: var(--border-radius);
		color: var(--input-text);
		background-color: var(--input-background-color);
	}

	.hidden {
		display: none;
	}

	.toggle-btn {
		position: relative;
		right: 0.6rem;
		top: 0;
		background: transparent;
		border: none;
		cursor: pointer;
		display: block;
	}
</style>
