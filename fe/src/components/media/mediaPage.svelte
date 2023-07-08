<script>
  import MediaCard from "./MediaCard.svelte";
  import ReviewList from "./ReviewList.svelte";

  export const id = "";
  export let title = "";
  export let image = "";
  export let averageRating = 0;
  export let creators = [];
  export let optionalField1 = null;
  export let optionalField2 = null;
  export let reviews = [];
</script>

<div class="media-page">
  <div class="media-page-content">
    <div class="media-info">
      <MediaCard
        {title}
        {image}
        {averageRating}
        {creators}
        {optionalField1}
        {optionalField2}
      />
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
