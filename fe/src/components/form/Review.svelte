<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { Track, Album } from '$lib/types/music.ts';
	import type { Keyword } from '$lib/types/media.ts';
	import type { UUID } from '$lib/types/utils.ts';
	import type { Review } from '$lib/types/review.ts';
	import { videoWork, isVideoWork } from '$stores/media/isVideo.ts';
	import { reviewStore } from '$stores/form/review.ts';
	import { keywordStore } from '$stores/form/keyword.ts';
	import { trackStore } from '$stores/media/music.ts';

	let album: Album;
	let mediaID: UUID = ''; // updated on fetch
	let isMediaVideo = false;
	let favoriteTrack = null;
	let tracks: Track[] = [];

	// subscribing to trackStore allows fetching user's favorite track etc.
	reviewStore.subscribe((value) => {
		favoriteTrack = value.favoriteTrack;
	});

	trackStore.subscribe((value) => {
		favoriteTrack = tracks.find((track) => track.media_id === value?.track?.media_id) || null;
	});

	const submitReview = async (mediaID: UUID) => {
		try {
			await reviewStore.submitReview(mediaID);
		} catch (error) {
			console.error(error);
		}
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		submitReview(mediaID);
	};

	function incrementVote(keyword: Keyword) {
		keywordStore.incrementVote(keyword);
	}

	function decrementVote(keyword: Keyword) {
		keywordStore.decrementVote(keyword);
	}

	async function suggestKeywords(mediaID: UUID) {
		await keywordStore.suggestKeywords(mediaID, $keywordStore.keywordSearch);
		keywordStore.clearSearch();
	}

	const handleFavoriteTrackChange = (e: Event) => {
		const selectedTrackId = (e.target as HTMLSelectElement).value;
		const selectedTrack = tracks.find((track) => track.media_id === selectedTrackId);

		selectedTrack
			? reviewStore.setFavoriteTrack(selectedTrack)
			: console.error('Selected track not found in trackStore');
	};

	const unsubscribe = videoWork.subscribe((value) => {
		isMediaVideo = value;
	});

	onMount(async () => {
		let response = await fetch(`/api/reviews/${mediaID}`);
		let reviews = await response.json();
		isVideoWork(mediaID);
	});

	onDestroy(() => {
		unsubscribe();
	});
</script>

<form on:submit={handleSubmit}>
	<h2>Review</h2>
	{#if isMediaVideo}
		<select bind:value={$reviewStore.favoriteTrack} on:input={handleFavoriteTrackChange}>
			<option value="">Select a favorite track</option>
			{#each album.tracks as track}
				<option value={track}>{track}</option>
			{/each}
		</select>
	{/if}

	<label>
		{'Track ratings'}
		{#if $reviewStore.trackRatings}
			{#each $reviewStore.trackRatings as trackRating, i (i)}
				<input
					type="number"
					bind:value={$reviewStore.trackRatings[i]}
					min="1"
					max={$reviewStore.ratingScale}
					required
				/>
			{/each}
		{:else}
			<p>No track ratings available</p>
		{/if}
	</label>

	{#if isMediaVideo}
		<label>
			<!-- TODO: refactor to accomodate new rating system -->
			Cast ratings
			{#if $reviewStore.castRatings}
				{#each $reviewStore.castRatings as castRating, i (i)}
					<input
						type="number"
						bind:value={castRating}
						min="1"
						max={$reviewStore.ratingScale}
						required
					/>
				{/each}
			{:else}
				<p>No cast ratings available</p>
			{/if}
		</label>
	{/if}

	<section class="expandable-box" role="region" aria-labelledby="keywords-heading">
		<h3>Keywords</h3>
		<ul>
			{#each $keywordStore.selectedKeywords as keyword (keyword.keyword)}
				<li>
					<span>{keyword.keyword} ({keyword.stars}/{$reviewStore.ratingScale})</span>
					<button
						type="button"
						on:click={() => incrementVote(keyword)}
						aria-label={`Increment vote for ${keyword.keyword}`}>+</button
					>
					<button
						type="button"
						on:click={() => decrementVote(keyword)}
						aria-label={`Decrement vote for ${keyword.keyword}`}>-</button
					>
				</li>
			{/each}
		</ul>

		<div class="keyword-search">
			<input
				type="text"
				bind:value={$keywordStore.keywordSearch}
				placeholder="Search keywords..."
				list="keywords"
			/>
			<datalist id="keywords">
				{#each $keywordStore.keywords as keyword (keyword)}
					<option value={keyword} />
				{/each}
			</datalist>
			<button type="button" on:click={() => suggestKeywords(mediaID)} aria-label="Suggest keywords"
				>&gt;&gt;</button
			>
		</div>
	</section>
	<label>
		<!-- TODO: allow setting custom values for min and max -->
		<textarea
			aria-label="Review text"
			bind:value={$reviewStore.reviewText}
			on:input={reviewStore.handleReviewChange}
			required
		/>
	</label>

	<div aria-live="polite">Word count: {$reviewStore.wordCount}</div>

	<button type="submit">Submit Review</button>
</form>

<!-- TODO: add CSS variables to ease theming -->
<style>
	:root {
		--input-border-color: #ccc;
		--input-border-color-focus: #aaa;
		--input-border-color-error: #f00;
		--input-background-color: #fff;
		--input-background-color-focus: #fff;
		--input-text: #000;
		--border-radius: 2px;
		--box-padding: 0.5em;
		--box-marigin-btm: 0.5em;
	}

	input,
	select,
	textarea,
	button {
		font-family: inherit;
		font-size: inherit;
		-webkit-padding: 0.4em 0;
		padding: 0.4em;
		margin: 0 0 0.5em 0;
		box-sizing: border-box;
		border: 1px solid #ccc;
		border-radius: 4px;
	}

	.expandable-box {
		border: 1px solid #ccc;
		padding: var(--box-padding);
		margin-bottom: var(--box-margin-btm);
	}

	.expandable-box h3 {
		margin-top: 0;
	}

	.expandable-box ul {
		list-style-type: none;
		padding-left: 0;
	}

	.expandable-box li {
		display: flex;
		align-items: center;
	}

	.expandable-box li span {
		flex-grow: 1;
	}

	.expandable-box li button {
		margin-left: 10px;
	}

	.keyword-search {
		display: flex;
		margin-top: 10px;
	}

	.keyword-search input {
		flex-grow: 1;
	}
</style>
