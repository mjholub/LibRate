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
		<span class="icon">
			{#if showPassword}
				<EyeIcon />
			{:else}
				<EyeOffIcon />
			{/if}
		</span>
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
		--border-radius: 4px;
		--pwd-container-display: inline-block;
	}

	.password-container {
		position: relative;
		overflow: visible;
		display: var(--pwd-container-display) !important;
		border: 1px solid var(--input-border-color);
		border-radius: var(--border-radius);
		color: var(--input-text);
		background-color: var(--input-background-color);
		width: calc(98% - 0.2em);
		height: 2rem;
	}

	input#password {
		font-family: inherit;
		font-size: inherit;
		-webkit-padding: 0.4em 0;
		padding: 0.2em 0 0.2em 0;
		margin: 0.1em 0 0.1em 0;
		box-sizing: border-box;
		border: none;
		border-radius: 4px;
		width: inherit;
		display: inline-block;
		position: relative;
	}

	.icon {
		/* any emoji or icon font */
		font-family: 'Noto Color Emoji', 'Material Icons', sans-serif;
		font-weight: normal;
		font-style: normal;
		font-size: 1.2rem; /* Preferred icon size */
		position: absolute;
		display: inline-block;
		line-height: 1;
		right: -0.4rem;
		bottom: -50%;
		text-transform: none;
		letter-spacing: normal;
		word-wrap: normal;
		white-space: nowrap;
		direction: ltr;
		-webkit-font-smoothing: antialiased;
		text-rendering: optimizeLegibility;
		-moz-osx-font-smoothing: grayscale;
		font-feature-settings: 'liga';
	}
	.hidden {
		display: none;
	}

	.toggle-btn {
		position: relative;
		right: 0.6rem;
		top: 30%;
		background: transparent;
		border: none;
		cursor: pointer;
		display: ruby-text-container;
	}
</style>
