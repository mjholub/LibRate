<script>
  import ReviewForm from "../form/Review.svelte";
  import ReviewCard from "./ReviewCard.svelte";

  export let reviews = [];
</script>

<div class="review-list">
  <ReviewForm />

  {#if reviews.length > 0}
    <div>
      {#each reviews as review (review.id)}
        <ReviewCard {review} />
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
