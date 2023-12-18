<script lang="ts">
	import Search from '$components/utility/Search.svelte';
	import ProfileControls from './ProfileControls.svelte';

	export let authenticated: boolean;
	export let nickname: string;
</script>

<div class="header">
	{#if authenticated}
		<!-- leave space for profile controls -->
		<span class="search-with-space">
			<Search />
		</span>
		<span class="profile-controls">
			<ProfileControls {nickname} />
		</span>
	{:else}
		<span class="search">
			<Search />
		</span>
	{/if}
</div>

<style lang="scss">
	$justify-header-content: left;

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.2em 0.6em 0 0;
		position: relative;
		z-index: 0;
	}

	.header > * {
		height: 2.5em;
		padding: 0.5em;
		flex-direction: row;
		display: grid;
		transition: transform 500ms ease-in-out;
		justify-content: $justify-header-content;
	}

	span.search-with-space {
		flex-grow: 3;
		min-width: 30%;
	}

	span.profile-controls {
		display: inline-flex;
		flex-shrink: 0;
		padding-left: 3.2em;
		align-items: center;
		justify-content: center;
		padding-block-end: 0.5em;
	}
</style>
