<script lang="ts">
	import type { TVShow, Episode } from '$lib/types/film_tv';
	import { Label, Input, Collapse } from '@sveltestrap/sveltestrap';

	let synopsisLanguage: string = 'en';

	let eps: Episode[] = [
		{
			season: 0,
			number: 0,
			title: '',
			plot: {
				String: '',
				Valid: false
			},
			duration: null,
			castID: 0,
			airDate: null,
			languages: []
		}
	];

	let show: TVShow = {
		title: '',
		seasons: [
			{
				number: 0,
				episodes: eps
			}
		]
	};

	const removeSeason = (event: any) => {
		if (event.target.value === '' && show.seasons.length > 1) {
			const indexToRemove = show.seasons.findIndex(
				(season) => season.number === event.target.value
			);
			show.seasons = show.seasons.filter((season, index) => index !== indexToRemove);
		}
	};

	const addSeason = (event: any) => {
		const newSeason = {
			number: show.seasons.length,
			episodes: []
		};
		show.seasons = [...show.seasons, newSeason];
	};

  const handleKeydown = (event: any) => {
    switch (event.key) {
      case 'Enter':
        addSeason(event);
        break;
      case 'Backspace':
        removeSeason(event);
        break;
    }
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css"
	/>
</svelte:head>

<form class="sub_form">
	<p class="warning">TV Show submission not working yet, this is just a mockup.</p>
	<Label for="title">Title</Label>
	<Input type="text" id="title" bind:value={show.title} />
	<!-- collapsible section for adding cast -->
	<Collapse
		>click here to start adding cast
		<Label for="cast">Cast</Label>
	</Collapse>
	<Label for="synopsis">Synopsis</Label>
	<button id="addSeason">Add Season</button>
	<!-- do someting like in tracks form -->
	<input type="text" id="season" on:keydown={handleKeydown} />
	{#each show.seasons as season}
		{#each season.episodes as episode}
			<Input
				type="textarea"
				id="synopsis"
				bind:value={episode.plot.String}
				on:input={() => {
					episode.plot.Valid = episode.plot.String.length > 5;
				}}
			/>
			<select id="lang" bind:value={synopsisLanguage}>
				<option value="en">English</option>
				<option value="fr">French</option>
			</select>
			<Input type="date" id="releaseDate" bind:value={episode.airDate} />
			<Label for="duration">Duration</Label>
			<Input type="number" id="duration" bind:value={episode.duration} />
		{/each}
	{/each}
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
