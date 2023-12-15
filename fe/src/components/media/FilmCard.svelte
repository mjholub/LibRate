<script lang="ts">
	/* TODO: Add stills in a gallery */
	import type { Film } from '$lib/types/film_tv.ts';
	console.info('mounting FilmCard initialized');
	export let posterPath: string;
	export let film: Film;

	let durationStr = '';
	let releaseDate = 'unknown';
	if (film.duration && film.duration.Valid) {
		durationStr = film.duration.Time.split('T')[1].split('.')[0];
	}
	if (film.releaseDate && film.releaseDate.Valid) {
		// cut to just the date
		releaseDate = film.releaseDate.Time.toDateString();
	}
</script>

{#if posterPath}
	<img class="media-image" src={posterPath} alt={film.title} />
{/if}

<dl class="film-details">
	<dt>Title:</dt>
	<dd>{film.title}</dd>
	<dt>Release date</dt>
	<dd>{releaseDate}</dd>
	<dt>Duration:</dt>
	<dd>{durationStr}</dd>
	<dt class="synopsis">Synopsis:</dt>
	<dd>{film.synopsis?.String}</dd>
</dl>

<style>
	:root {
		--film-card-width: 30vw;
		--film-card-height: 50vh;
	}
	.media-image {
		display: block;
		width: var(--film-card-width);
		height: var(--film-card-height);
	}
	.film-details {
		display: grid;
		grid-template-columns: 1fr 1fr;
		grid-template-rows: 1fr 1fr;
		width: var(--film-card-width);
		height: var(--film-card-height);
		margin: 0;
		padding: 0;
	}
	dt {
		font-weight: bold;
	}
</style>
