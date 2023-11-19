<script lang="ts">
	import { onMount } from 'svelte';
	import ReviewForm from '../form/Review.svelte';
	import ReviewCard from './ReviewCard.svelte';
	import type { Review } from '$lib/types/review.ts';

	export let reviews: Review[];
	const limit = 5;
	let offset = 0;
	const nick = 'anonymous'; // FIXME: add nick fetching logic here

	// TODO: move to store
	const getReviews = async (limit: number, offset: number) => {
		const params = new URLSearchParams({
			limit: limit.toString(),
			offset: offset.toString()
		});
		const res = await fetch('/api/reviews/latest?' + params.toString(), {
			method: 'GET',
			headers: {
				'Content-Type': 'application/json'
			}
		});
		const data = await res.json();
		reviews = data;
	};

	const loadMore = async () => {
		offset += limit;
		const params = new URLSearchParams({
			limit: limit.toString(),
			offset: offset.toString()
		});
		const res = await fetch('/api/reviews/latest?' + params.toString(), {
			method: 'GET',
			headers: {
				'Content-Type': 'application/json'
			}
		});
		const data = await res.json();
		reviews = [...reviews, ...data];
	};

	onMount(async () => {
		getReviews(limit, offset);
	});
</script>

<div class="review-list">
	<ReviewForm />

	{#if reviews.length > 0}
		<div>
			{#each reviews as review (review.id)}
				<ReviewCard {review} userid={review.userid} {nick} media={review.media} />
			{/each}
			<button on:click={loadMore}>Load More Reviews</button>
		</div>
	{:else}
		<div class="no-reviews">No reviews yet. Be the first to write one!</div>
	{/if}
</div>

<style>
	:root {
		--no-reviews-color: #666;
	}

	.review-list {
		margin-top: 2em;
	}

	.no-reviews {
		text-align: center;
		color: var(--no-reviews-color);
	}
</style>
