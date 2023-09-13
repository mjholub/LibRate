<script lang="ts">
	/* TODO: Add stills in a gallery */
	import type { Film } from '$lib/types/film_tv.ts';
	console.info('mounting FilmCard initialized');
	export let posterPath: string;
	export let film: Film;
</script>

{#if posterPath}
	<img class="media-image" src={posterPath} alt={film.title} />
{/if}

<div class="film-details">
	<div><b>Title:</b> {film.title}</div>
	<div><b>Release Date</b>: {film.releaseDate}</div>
	<div><b>Duration</b>: {film.duration}</div>
	<div><b>Rating</b>: {film.rating}</div>
</div>
