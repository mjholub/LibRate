<script lang="ts">
	import { onDestroy } from 'svelte';
	import { submitMediaForm } from '$stores/form/add_media';
	import type { Media } from '$lib/types/media';
	import AddAlbum from './media/AddAlbum.svelte';

	let media: Media;

	const unsubscribe = submitMediaForm.subscribe((value) => {
		media = value;
	});

	submitMediaForm.subscribe((value) => {
		media = value;
	});

	onDestroy(() => {
		unsubscribe();
		submitMediaForm.set(media);
	});

	// return value doesn't matter here
	const submitMedia = async (media: Media) => {
		try {
			await submitMediaForm.submitMedia(media);
		} catch (error) {
			console.error(error);
		}
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		submitMedia(media);
	};
</script>

<div class="form-body">
	<h2 class="form-title">Add Media</h2>
	<form on:submit|preventDefault={handleSubmit}>
		<label for="kind">Select media type:</label>
		<select bind:value={media.kind} id="kind">
			<option value="none">---</option>
			<option value="album">Album</option>
			<option value="film">Film</option>
			<option value="tv_show">TV Show</option>
			<option value="book">Book</option>
			<option value="anime">Anime</option>
			<option value="manga">Manga</option>
			<option value="comic">Comic</option>
			<option value="game">Game</option>
		</select>

		{#if media.kind == 'none'}
			<p>Please select a media type.</p>
		{/if}

		{#if media.kind === 'album'}
			<AddAlbum />
		{/if}

		{#if media.kind === 'film'}
			<AddFilm />
		{/if}

		{#if media.kind === 'tv_show'}
			<AddTVShow />
		{/if}

		{#if media.kind === 'book'}
			<AddBook />
		{/if}

		{#if media.kind === 'anime'}
			<AddAnime />
		{/if}

		{#if media.kind === 'manga'}
			<AddManga />
		{/if}

		{#if media.kind === 'comic'}
			<AddComic />
		{/if}

		{#if media.kind === 'game'}
			<AddGame />
		{/if}
	</form>
</div>
