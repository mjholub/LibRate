<script lang="ts">
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
		aria-label="Password"
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
		<span class="material-icons">
			{showPassword ? 'visibility_off' : 'visibility'}
		</span>
	</button>
</div>

<style>
	.password-container {
		position: relative;
		overflow: hidden;
		display: flex;
	}

	.hidden {
		display: none;
	}

	.toggle-btn {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		background: transparent;
		border: none;
		cursor: pointer;
	}

	.material-icons {
		font-family: 'Material Icons';
		font-weight: normal;
		font-style: normal;
		font-size: 20px; /* Preferred icon size */
		display: inline-block;
		line-height: 1;
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
</style>
