<script lang="ts">
	import { PlusIcon, XIcon } from 'svelte-feather-icons';
	import type { Album } from '$lib/types/music';
	import type { Genre } from '$lib/types/media';
	let album: Album = {
		UUID: '',
		kind: 'album',
		image_paths: [],
		media_id: '',
		name: '',
		title: '',
		created: new Date(),
		creator: null,
		creators: [],
		added: new Date(),
		album_artists: {
			person_artist: [],
			group_artist: []
		},
		release_date: '',
		genres: [],
		duration: {
			Valid: false,
			Time: '00:00:00'
		},
		tracks: []
	};

	const addMore = () => {};
	const addImage = () => {};
	const removeGenre = (index: number) => {
		if (album.genres) {
			album.genres.splice(index, 1);
		}
	};
	let genres: string = '';
</script>

<div
	class="drop-area"
	on:drop={addImage}
	on:dragover={(e) => e.preventDefault()}
	aria-dropeffect="copy"
	role="region"
	aria-labelledby="drop-area-label"
>
	<p id="drop-area-label">Drop or click to add album cover here</p>
	{#if album.image_paths}
		<img src={album.image_paths[0]} alt="Album Cover" />
	{/if}
</div>

<label for="name">Album Name:</label>
<input id="name" bind:value={album.name} />

<label for="album-artists">Album Artists:</label>
<select id="album-artists" bind:value={album.album_artists}>
	<option value="person_artist">Person</option>
	<option value="group_artist">Group</option>
</select>

<button id="add-more" on:click={addMore}>
	<PlusIcon />
</button>

<label for="release-date">Release Date:</label>
<input id="release-date" bind:value={album.release_date} type="date" />

<label for="genres">Genres (comma separated):</label>
<div>
	{#if album.genres}
		{#each album.genres as genre, index}
			<div class="genre-box">
				{genre}
				<span
					class="remove-genre"
					on:click={() => removeGenre(index)}
					on:keyup={(e) => e.key === 'Enter' && removeGenre(index)}
					aria-label="Remove genre"
					role="button"
					tabindex="0"
				>
					<XIcon size="12" />
				</span>
			</div>
		{/each}
	{/if}
</div>
<input id="genres" bind:value={genres} on:blur={() => (album.genres = genres.split(','))} />

<label for="duration">Duration:</label>
<input id="duration" bind:value={album.duration} type="time" />

<p>Tracks:</p>
<p>Tracks:</p>

<style>
	.drop-area {
		border: 2px dashed #ccc;
		padding: 20px;
		text-align: center;
	}

	.drop-area img {
		max-width: 100%;
		max-height: 200px;
		margin-top: 10px;
	}

	.genre-box {
		display: inline-block;
		margin: 0 8px 8px 0;
		padding: 6px 12px;
		background-color: #f0f0f0;
		border: 1px solid #ccc;
		border-radius: 4px;
		position: relative;
	}

	.remove-genre {
		cursor: pointer;
		position: absolute;
		top: 50%;
		right: 8px;
		transform: translateY(-50%);
		color: #888;
	}
</style>
