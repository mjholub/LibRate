<script lang="ts">
	import type { Album } from '../../types/music.ts';
	//import { mediaImageStore } from '../../stores/media/image.ts';
	let showArtists = false;
	let durationStr = '';
	function toggleArtists() {
		showArtists = !showArtists;
	}
	export let album: Album;
	if (album.duration && album.duration.Valid) {
		durationStr = album.duration.Time.split('T')[1].split('.')[0];
	}
	export let imgPath: string;
	console.info('mounting AlbumCard initialized');
</script>

<!-- WARN: optimistically assuming the first image is the album cover -->
{#if imgPath}
	<img class="media-image" src={imgPath} alt={album.name} />
{/if}
<div class="album-details">
	<div><b>Title:</b> {album.title}</div>
	<!-- merge the artists into one array -->
	<div>
		<b>Artists:</b>
		{#if album.album_artists.group_artist !== undefined}
			{#each album.album_artists.group_artist as artist}
				{artist.name}
			{/each}
		{/if}
		<!-- expandable list of artists from the person_artist -->
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
	</div>
	<div>Release Date: {album.release_date}</div>
	<div class="album-tracklist">
		{#each album.tracks as track (track.media_id)}
			<p>Tracklist:</p>
			<div class="album-track">
				<span>{track.track_number}</span>
				<span>{track.name}</span>
				<span>{track.duration}</span>
			</div>
		{/each}
	</div>
	<div>Duration: {durationStr}</div>
</div>

<style>
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
		background-color: #4caf50; /* Green */
		border: none;
		color: white;
		padding: 0.2em 0.3em;
		text-align: center;
		text-decoration: none;
		display: inline-block;
		font-size: 15px;
		margin: 4px 2px;
		cursor: pointer;
		border-radius: 4px;
		transition-duration: 0.4s;
	}

	.toggle-button:hover {
		background-color: #45a049;
	}
</style>
