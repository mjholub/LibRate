<script lang="ts">
	import MediaCard from './MediaCard.svelte';
	import ReviewList from '../review/ReviewList.svelte';
	import type { Person } from '../../types/people.ts';
	import type { Review } from '../../types/review.ts';

	export const id = '';
	export let title = '';
	export let image = '';
	export let averageRating = 0;
	export let creators: Person[] = [];
	export let reviews: Review[] = [];
</script>

<div class="media-page">
	<div class="media-page-content">
		<div class="media-info">
			<MediaCard {title} {image} {averageRating} {creators} />
		</div>

		<div class="reviews">
			<ReviewList {reviews} />
		</div>
	</div>
</div>

<style>
	.media-page {
		padding: 1em;
	}

	.media-page-content {
		display: flex;
		gap: 2em;
	}

	.media-info {
		flex: 1 0 40%;
	}

	.reviews {
		flex: 1 0 60%;
	}
</style>
