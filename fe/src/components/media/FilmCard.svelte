<script lang="ts">
	/* TODO: Add stills in a gallery */
	import type { Film } from '$lib/types/film_tv.ts';
	console.info('mounting FilmCard initialized');
	export let posterPath: string;
	export let film: Film;

	let durationStr = '';
	if (film.duration && film.duration.Valid) {
		durationStr = film.duration.Time.split('T')[1].split('.')[0];
	}
</script>

{#if posterPath}
	<img class="media-image" src={posterPath} alt={film.title} />
{/if}

<dl class="film-details">
	<dt>Title:</dt>
	<dd>{film.title}</dd>
	<dt>Release date</dt>
	<dd>{film.releaseDate}</dd>
	<dt>Duration:</dt>
	<dd>{durationStr}</dd>
	<dt class="synopsis">Synopsis</dt>
	<dd>{film.synopsis}</dd>
</dl>
