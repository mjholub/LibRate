<script lang="ts">
	import type { Film } from '$lib/types/film_tv';
	import { Label, Input, Collapse } from '@sveltestrap/sveltestrap';

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

<form class="sub_form">
	<p class="warning">Film submission not working yet, this is just a mockup.</p>
	<Label for="title">Title</Label>
	<Input type="text" id="title" bind:value={film.title} />
	<!-- collapsible section for adding cast -->
	<Collapse
		>click here to start adding cast
		<Label for="cast">Cast</Label>
	</Collapse>
	<Label for="synopsis">Synopsis</Label>
	<Input
		type="textarea"
		id="synopsis"
		bind:value={film.synopsis.String}
		on:input={() => {
			film.synopsis.Valid = film.synopsis.String.length > 5;
		}}
	/>
	<select id="lang" bind:value={synopsisLanguage}>
		<option value="en">English</option>
		<option value="fr">French</option>
	</select>
	<Input type="date" id="releaseDate" bind:value={film.releaseDate.Time} />
	<Label for="duration">Duration</Label>
	<Input type="number" id="duration" bind:value={film.duration} />
	<button type="submit">Submit</button>
</form>

<style>
	.sub_form {
		display: block;
	}
	label {
		display: block;
		margin-top: 1rem;
	}
	input {
		margin-top: 0.4rem;
		margin-bottom: 0.2rem;
	}

	.warning {
		color: red;
		font-weight: bold;
		font-size: 1.5em;
	}
</style>
