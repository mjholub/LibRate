<script lang="ts">
	import type { Album } from '$lib/types/music.ts';
	let showArtists = false;
	let showKeywords = false;
	let durationStr = '';
	function toggleArtists() {
		showArtists = !showArtists;
	}
	function toggleKeywords() {
		showKeywords = !showKeywords;
	}
	export let album: Album;
	if (album.duration && album.duration.Valid) {
		const durationDate = new Date(album.duration.Time);
		// FIXME: ditch static adapter, because without
		// the ability to use dynamic rendering, this will always
		// be the host machine's locale, while we want it to be the user's
		// browser locale
		durationStr = durationDate.toLocaleTimeString(undefined, {
			hour: 'numeric',
			minute: 'numeric',
			second: 'numeric'
		});
	}
	export let imgPath: string;
	console.info('mounting AlbumCard initialized');
</script>

<!-- WARN: optimistically assuming the first image is the album cover -->
{#if imgPath}
	<img class="media-image" src={imgPath} alt={album.name} loading="lazy" />
{/if}
<div class="album-details">
	<dl>
		<dt>Title:</dt>
		<dd>{album.title}</dd>
		<!-- merge the artists into one array -->
		<dt>Artists:</dt>
		<dd>
			{#if album.album_artists.group_artist !== undefined}
				{#each album.album_artists.group_artist as artist}
					{artist.name}
				{/each}
			{/if}
			<button class="toggle-button" on:click={toggleArtists}>
				{#if showArtists}
					Show less
				{:else}
					Show more
				{/if}
			</button>
			{#if showArtists}
				<div>
					{#each album.album_artists.person_artist as artist}
						{artist.first_name}
						{#if artist.nick_names}"{artist.nick_names}"{/if}
						{artist.last_name}
					{/each}
				</div>
			{/if}
		</dd>

		<dt>Release Date:</dt>
		<dd>{album.release_date}</dd>

		<dt>Tracklist:</dt>
		<dd class="album-tracklist">
			{#each album.tracks as track (track.media_id)}
				<div class="album-track">
					<span>{track.track_number++}</span>
					<span>{track.name}</span>
					<span>{track.duration}</span>
				</div>
			{/each}
		</dd>

		<dt>Duration:</dt>
		<dd>{durationStr}</dd>
		<!-- TODO: ability to add keywords from this component-->

		<dt>Keywords:</dt>
		<dd class="keywords-container">
			<div class="keywords">
				<button class="toggle-button" on:click={toggleKeywords}>
					{#if showKeywords}
						Hide
					{:else}
						Show
					{/if}
				</button>
			</div>
			{#if showKeywords}
				{#if album.keywords !== undefined}
					{#each album.keywords as keyword}
						<span>{keyword.keyword}</span>
					{/each}
				{/if}
			{/if}
		</dd>
	</dl>
</div>

<style>
	/* TODO: use more CSS variables */
	:root {
		--input-border-color: #ccc;
		--input-border-color-focus: #aaa;
		--input-border-color-error: #f00;
		--input-background-color: #fff;
		--input-background-color-focus: #fff;
		--input-text: #000;
		--border-radius: 2px;
		--toggle-btn-bgcolor: #4caf50;
		--toggle-btn-hover-bgcolor: #45a049;
	}

	.album-details {
		display: flex;
		width: 100%;
		flex-direction: column;
	}
	.album-tracklist {
		display: flex;
		flex-direction: column;
		width: 100%;
		overflow-x: auto;
	}
	.album-track {
		display: flex;
		width: 100%;
		justify-content: space-between;
	}
	.toggle-button {
		background-color: var(--toggle-btn-bgcolor); /* Green */
		border: none;
		color: white;
		padding: 0.2em 0.3em;
		text-align: center;
		text-decoration: none;
		display: inline-block;
		font-size: 0.8em;
		margin: 0.2em 0.1em;
		cursor: pointer;
		border-radius: 4px;
		transition-duration: 0.4s;
	}

	.toggle-button:hover {
		background-color: var(--toggle-btn-hover-bgcolor);
	}
	.keywords-container {
		display: flex;
		flex-direction: row;
		font-size: 0.8em;
		/* pale gray */
		color: #d3d3d3;
	}
	.keywords {
		display: flex;
		flex-direction: row;
		font-size: 1em;
		color: #ffffff;
	}
</style>
