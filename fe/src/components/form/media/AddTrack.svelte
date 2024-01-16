<script lang="ts">
	import type { Track } from '$lib/types/music';

	let tracks: Track[] = [];

	import DragDropList from '$components/shared/DragDropList.svelte';
</script>

<div>
	<DragDropList bind:data={tracks} removesItems={true} />
</div>

<style>
	/* Add styles for AddTrack.svelte */
</style>
