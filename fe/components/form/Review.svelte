<script>
  import { onMount } from "svelte";

  let favoriteTrack = "";
  let trackRatings = "";
  let castRatings = "";
  let themeVotes = [];
  let reviewText = "";
  let wordCount = 0;
  let ratingScale = 10; // Default rating scale
  let mediaID = 0; // updated on fetch

  onMount(async () => {
    // TODO: Replace this with actual fetching of user preference
    ratingScale = await fetchUserRatingPreference();
    let response = await fetch("/api/reviews/${mediaID}");
    let review = await response.json();
  });

  const submitReview = async () => {
    if (wordCount < 20) {
      alert("Review must be at least 20 words!");
      return;
    }

    let memberID = 1; // Replace with actual member ID
    let response = await fetch("/api/reviews", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        MemberID: memberID,
        MediaID: 1, // Replace with the actual media ID
        ReviewText: reviewText,
      }),
    });

    if (response.ok) {
      alert("Review submitted successfully!");
    } else {
      alert("Failed to submit review.");
    }
  };

  const handleReviewChange = (event) => {
    reviewText = event.target.value;
    wordCount = reviewText.split(/\s+/).length;
  };

  const fetchUserRatingPreference = () => {
    // Fetch user preference logic here...
    return Promise.resolve(10);
  };
</script>

<form on:submit|preventDefault={submitReview}>
  <label>
    Favorite track
    <input type="text" bind:value={favoriteTrack} required />
  </label>

  <label>
    Track ratings
    <input
      type="number"
      bind:value={trackRatings}
      min="1"
      max={ratingScale}
      required
    />
  </label>

  <label>
    Cast ratings
    <input
      type="number"
      bind:value={castRatings}
      min="1"
      max={ratingScale}
      required
    />
  </label>

  <label>
    Theme votes
    <input
      type="number"
      bind:value={themeVotes}
      min="0"
      max={ratingScale}
      required
    />
  </label>

  <label>
    Review (min 20 words)
    <textarea bind:value={reviewText} on:input={handleReviewChange} required />
  </label>

  <div>Word count: {wordCount}</div>

  <button type="submit">Submit Review</button>
</form>
