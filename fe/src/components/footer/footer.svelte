<script lang="ts">
	import { onMount } from 'svelte';
	let version = 'Loading...';

	onMount(async () => {
		const res = await fetch('/package.json');
		const pkg = await res.json();
		version = pkg.version;
	});
</script>

<footer>
	<p>LibRate v{version}</p>
	<p>
		Copyright 2023 <a href="https://codeberg.org/mjh/LibRate/src/branch/main/CONTRIBUTORS.md"
			>LibRate contributors</a
		>
	</p>
	<p>Licensed under the <a href="https://www.gnu.org/licenses/agpl-3.0.en.html">AGPL-3.0</a></p>
</footer>

<style>
	footer {
		text-align: center;
		padding: 1rem;
		font-size: 0.8rem;
		color: #fff;
		background-color: #333;
	}
</style>
