<script lang="ts">
	import { EyeIcon, EyeOffIcon } from 'svelte-feather-icons';

	export let value: string;
	export let id: string;
	export let onInput: (password: string) => Promise<void>;
	export let showPassword: boolean;
	export let toggleObfuscation: () => void;
</script>

<div class="password-container">
	<!-- bind the value to the input, then re-emit the input event -->
	<!-- If we need to then assign an input to a function that accepts some params
  we need to wrap it in a higher order function -->
	{#if !showPassword}
		<input
			{id}
			bind:value
			type="password"
			autocomplete="new-password"
			aria-label="Password"
			aria-live="polite"
			required
			on:input={() => {
				onInput(value);
			}}
		/>
	{:else}
		<input
			id="{id}Text"
			bind:value
			type="text"
			aria-live="polite"
			autocomplete="new-password"
			required
			aria-label="Password (visible)"
		/>
	{/if}
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
	}

	.password-container {
		position: relative;
		overflow: visible;
		display: inline-grid;
		border: none;
		border-radius: var(--border-radius);
		color: var(--input-text);
		width: 97%;
		height: 2rem;
	}

	input#password,
	input#passwordText {
		font-family: inherit;
		font-size: inherit;
		margin: 0.1em 0 0.1em 0;
		box-sizing: border-box;
		border-radius: 4px;
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
		right: 6%;
		bottom: 0;
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

	.toggle-btn {
		position: abosolute !important;
		right: 0.75rem;
		top: 50%;
		background: transparent;
		border: none;
		cursor: pointer;
		display: inline-grid;
	}
</style>
