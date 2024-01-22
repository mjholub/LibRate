<script lang="ts">
	import type { Film } from '$lib/types/film_tv';
	import {
		Card,
		Label,
		Input,
		FormGroup,
		ListGroup,
		Collapse,
		ListGroupItem
	} from '@sveltestrap/sveltestrap';

	let synopsisLanguage: string = 'en';

	let film: Film = {
		title: '',
		castID: 0,
		synopsis: {
			String: '',
			Valid: false
		},
		releaseDate: {
			Time: new Date(),
			Valid: false
		},
		duration: null,
		rating: 0
	};
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css"
	/>
</svelte:head>

<div class="film-meta">
	<FormGroup>
		<Label for="title">Title</Label>
		<Input type="text" id="title" bind:value={film.title} />
		<!-- collapsible section for adding cast -->
		<Collapse>click here to start adding case</Collapse>
		<Input type="textarea" id="synopsis" bind:value={film.synopsis.String} />
		<select id="lang" bind:value={synopsisLanguage}>
			<option value="en">English</option>
			<option value="fr">French</option>
		</select>
		<input type="date" id="releaseDate" bind:value={film.releaseDate.Time} />
		<input type="number" id="duration" bind:value={film.duration} />
	</FormGroup>
</div>

<style>
	/* Add styles for AddFilm.svelte */
</style>
