<script lang="ts">
	import MediaCard from './MediaCard.svelte';
	import type { Album } from '../../types/music.ts';

	export let album: Album;
	// TODO: create a store that'd fetch the main image for an album
	//export let albumMainImage: string;
</script>

<!-- WARN: optimistically assuming the first image is the album cover -->
{#if album.image_paths.length > 0}
	{#if album.image_paths[0] !== ''}
		<img class="media-image" src={album.image_paths[0]} alt={album.title} />
	{/if}
{/if}
<MediaCard
	media={album}
	image={album.image_paths[0]}
	title={album.title}
	creators={album.album_artists}
/>
<div class="album-details">
	<div>Title: {album.title}</div>
	<div>Release Date: {album.release_date}</div>
	<div class="album-tracklist">
		{#each album.tracks as track (track.media_id)}
			<tr>
				<td>{track.track_number}</td>
				<td>{track.title}</td>
				<td>{track.duration}</td>
			</tr>
		{/each}
	</div>
	<div>Duration: {album.duration}</div>
</div>
