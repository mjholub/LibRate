<script>
  import { onMount } from "svelte";

  export let albums = [];
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

  // TODO: check if this works
  const isVideoWork = async () => {
    let response = await fetch("/api/media/${mediaID}");
    let media = await response.json();
    return (
      media.Type === "Film" || media.Type === "TV" || media.Type === "Anime"
    );
  };

  (async () => {
    await Promise.all(
      albums.map(async (album) => {
        const response = await fetch(`/api/albums/${album.id}`);
        const albumData = await response.json();
        albums.push(albumData);
      })
    );
  })();
</script>

<form on:submit|preventDefault={submitReview}>
  <select bind:value={favoriteTrack}>
    <option value="">Select a favorite track</option>
    {#each albums as album}
      <option value={album.track}>{album.track}</option>
    {/each}
  </select>

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

  {#if isVideoWork}
    <label>
      Cast ratings
      <input
        type="number"
        bind:value={castRatings}
        min="1"
        max={ratingScale}
        placeholder="Cast ratings"
        required
      />
    </label>
  {/if}

  <label>
    <input
      type="number"
      bind:value={themeVotes}
      min="0"
      max={ratingScale}
      placeholder="Theme votes"
      required
    />
  </label>

  <label>
    <textarea
      bind:value={reviewText}
      on:input={handleReviewChange}
      placeholder="Review (min 20 words)"
      required
    />
  </label>

  <div>Word count: {wordCount}</div>

  <button type="submit">Submit Review</button>
</form>
