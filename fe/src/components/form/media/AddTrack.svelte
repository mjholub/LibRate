<script lang="ts">
	import { dndzone } from 'svelte-dnd-action';
	import { PlusIcon, XIcon } from 'svelte-feather-icons';
	import { flip } from 'svelte/animate';
	import type { Track } from '$lib/types/music';

	const flipDurationMs = 200;
	export let receivedAlbumID: string;

	function handleSort(e: CustomEvent) {
		tracks = e.detail.tracks.sort((a: any, b: any) => a.track_number - b.track_number);
	}
	let newInput: HTMLInputElement;
	let tracks: Track[] = [
		{
			track_number: 0,
			name: '',
			duration: 0,
			media_id: '',
			album_id: receivedAlbumID,
			lyrics: ''
		}
	];

	const removeElement = (event: any) => {
		if (event.target.value === '' && tracks.length > 1) {
			const indexToRemove = tracks.findIndex((track) => track.name === event.target.value);
			tracks = tracks.filter((track, index) => index !== indexToRemove);
		}
	};

	const addElement = (event: any) => {
		const newtrack: Track = {
			track_number: tracks.length,
			name: '',
			media_id: '',
			album_id: receivedAlbumID,
			duration: 0
		};
		tracks = [...tracks, newtrack];
		// @ts-ignore
		newInput = newtrack;
		setTimeout(() => {
			newInput.focus();
		}, 0);
	};

	const handleKeydown = (event: any) => {
		switch (event.key) {
			case 'Enter':
				addElement(event);
				break;
			case 'Backspace':
				removeElement(event);
				break;
		}
	};
	let items = tracks.map((track) => ({ id: track.track_number, item: track }));
</script>

<section use:dndzone={{ items, flipDurationMs }} on:consider={handleSort} on:finalize={handleSort}>
	{#each tracks as track (track.track_number)}
		<div animate:flip={{ duration: flipDurationMs }}>
			<span class="item-idx">#{track.track_number + 1}</span>
			<input
				bind:value={track.name}
				on:keydown={handleKeydown}
				bind:this={newInput}
				placeholder="Track {track.track_number + 1} title"
			/>
			<input
				id="duration-input"
				type="text"
				required
				pattern="[0-9]{2}:[0-9]{2}:[0-9]{2}"
				value="00:00:00"
			/>
			<span class="buttons-container">
				<button id="add-button" on:click={addElement}><PlusIcon /></button>
				<button id="del-button" on:click={removeElement}><XIcon /></button>
			</span>
		</div>
	{/each}
</section>

<style>
	section {
		width: 12em;
		padding: 1em;
	}
	div {
		height: 1.5em;
		width: 15em;
		text-align: center;
		border: 1px soltrack_number black;
		margin: 0.2em;
		padding: 0.3em;
		display: flex;
		width: fit-content;
		padding: 0.33em 0.7em 0.33em 0.4em;
	}
	.buttons-container {
		display: inline-flex;
		position: relative;
		flex-grow: 1;
		align-self: end;
		justify-content: center;
		right: 1%;
	}
	button {
		display: inline-flex;
		padding: 0 0.2em 0 0.2em;
		margin-right: 2%;
		position: sticky;
		margin-left: 6%;
	}
	button#add-button {
		margin-inline-start: 10%;
	}
	.item-idx {
		display: inline-block;
		margin-right: 2%;
		font-weight: 500;
	}
	input {
		display: flex;
		max-width: 80%;
		position: static;
		margin-right: 1.5%;
	}
	#duration-input {
		width: 100%;
		max-width: 25%;
		position: sticky;
		display: inline-flex;
	}
</style>
