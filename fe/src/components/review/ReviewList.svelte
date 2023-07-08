<script lang="ts">
	import { onMount } from 'svelte';
	import ReviewForm from '../form/Review.svelte';
	import ReviewCard from './ReviewCard.svelte';
	import type { Review } from '../../types/review.ts';

	export let reviews: Review[];

	const getReviews = async () => {
		const res = await fetch('/api/reviews');
		const data = await res.json();
		reviews = data;
	};

	onMount(() => {
		getReviews();
	});
</script>

<div class="review-list">
	<ReviewForm />

	{#if reviews.length > 0}
		<div>
			{#each reviews as review (review.id)}
				<ReviewCard {review} userid={review.userid} media={review.media} />
			{/each}
		</div>
	{:else}
		<div class="no-reviews">No reviews yet. Be the first to write one!</div>
	{/if}
</div>

<style>
	.review-list {
		margin-top: 2em;
	}

	.no-reviews {
		text-align: center;
		color: #666;
	}
</style>
