<script lang="ts">
	import { memberStore } from '$stores/members/getInfo.ts';
	import type { Review } from '$lib/types/review.ts';
	import type { Media } from '$lib/types/media.ts';
	import { keywordStore } from '$stores/form/keyword.ts';

	export let media: Media;
	export let review: Review;
	let userid: number;
	let userImage = '';
	export let nick: string;

	$: (async () => {
		if (userid) {
			userid = await memberStore.getMemberIDByNick(nick);
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

	{#if media.kind === 'film' || media.kind === 'tv_show' || media.kind === 'anime'}
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
	.review-card {
		border: 1px solid #ccc;
		padding: 1em;
		margin: 1em 0;
	}

	.review-user {
		display: flex;
		align-items: center;
		margin-bottom: 1em;
	}

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
