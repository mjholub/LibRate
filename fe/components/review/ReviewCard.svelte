<script>
  export let review = {
    username: "",
    userImage: "",
    text: "",
    date: "",
    rating: 0,
    favoriteTrack: "",
    trackRatings: [],
    castRatings: [],
    themeVoting: [],
  };
</script>

<div class="review-card">
  <div class="review-user">
    <img
      class="review-user-image"
      src={review.userImage}
      alt="{review.username}'s profile picture"
    />
    <div>{review.username}</div>
  </div>

  <div class="review-content">{review.text}</div>
  <div>Rating: {review.rating}</div>
  <div>Favorite track: {review.favoriteTrack}</div>

  <div>
    <strong>Track ratings:</strong>
    <ul>
      {#each review.trackRatings as trackRating (trackRating.id)}
        <li>{trackRating.track}: {trackRating.rating}</li>
      {/each}
    </ul>
  </div>

  <div>
    <strong>Cast ratings:</strong>
    <ul>
      {#each review.castRatings as castRating (castRating.id)}
        <li>{castRating.cast}: {castRating.rating}</li>
      {/each}
    </ul>
  </div>

  <div>
    <strong>Theme voting:</strong>
    <ul>
      {#each review.themeVoting as themeVote (themeVote.id)}
        <li>{themeVote.theme}: {themeVote.vote}</li>
      {/each}
    </ul>
  </div>

  <div>Reviewed on {review.date}</div>
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
