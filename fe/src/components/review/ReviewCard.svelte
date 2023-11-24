<script lang="ts">
	import { memberStore } from '$stores/members/getInfo.ts';
	import type { Review } from '$lib/types/review.ts';
	import type { Media } from '$lib/types/media.ts';
	import { keywordStore } from '$stores/form/keyword.ts';

	export let media: Media;
	export let review: Review;
	let userImage = '';
	export let nick: string;

	$: (async () => {
		if (nick) {
			userImage = await fetch('/static/images/' + nick + '.png').then((r) => r.text());
		}
	})();
</script>

<div class="review-card">
	<div class="review-user">
		<img class="review-user-image" src={userImage} alt="{nick}'s profile picture" />
		<div>{nick}</div>
	</div>

	<div class="review-content">{review.comment}</div>
	<div>Rating: {review.numstars}</div>
	<!--<div>Favorite track: {review.favoriteTrack}</div>-->
	<!-- FIXME: add this property to the review type and in the DB -->

	{#if media.kind === 'album'}
		<div>
			<strong>Track ratings:</strong>
			<ul>
				{#each review.trackratings as trackRating (trackRating.id)}
					<li>{trackRating.track}: {trackRating.rating}</li>
				{/each}
			</ul>
		</div>
	{/if}

	{#if ['film', 'tv_show', 'anime'].includes(media.kind)}
		<div>
			<strong>Cast rating:</strong>
			<ul>
				{#each review.castrating as castRating (castRating.id)}
					<li>{castRating.cast}: {castRating.numstars}</li>
				{/each}
			</ul>
		</div>
	{/if}

	<!-- Keywords voting -->
	<div>
		<strong>Vote for most relevant tags:</strong>
		<ul>
			{#each $keywordStore.keywords as keywordVote (keywordVote.id)}
				<li>{keywordVote.keyword}: {keywordVote.stars}</li>
			{/each}
		</ul>
	</div>

	<div>Reviewed on {review.created_at}</div>
</div>

<style>
	:root {
		--primary-color: #1a1a1a;
		--secondary-color: #e6e6e6;
		--tertiary-color: #ccc;
		--review-card-background-color: #fff;
		--review-card-border-style: solid;
		--review-card-border-color: #ccc;
		--review-card-padding: 0.15em;
		--review-card-margin: 0.4em 0;
	}

	.review-card {
		border: var(--review-card-border-style) var(--review-card-border-color);
		background-color: var(--review-card-background-color);
		padding: var(--review-card-padding);
		margin: var(--review-card-margin);
	}

	.review-user {
		display: flex;
		align-items: center;
		margin-bottom: 1em;
	}

	/* TODO: review if it handles different image sizes well when this becomes functional */
	.review-user-image {
		width: 50px;
		height: 50px;
		border-radius: 50%;
		margin-right: 1em;
	}

	.review-content {
		margin-bottom: 1em;
	}
</style>
