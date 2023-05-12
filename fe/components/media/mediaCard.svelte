<script>
  export let title = "";
  export let image = "";
  export let averageRating = 0;
  export let creators = [];
  export let optionalField1 = null; // FIXME: replace these with some real data
  export let optionalField2 = null;
</script>

<div class="media-card">
  <img class="media-image" src={image} alt={title} />
  <div class="media-title">{title}</div>
  <div>Average rating: {averageRating}</div>
  {#if optionalField1}
    <div>{optionalField1}</div>
  {/if}
  {#if optionalField2}
    <div>{optionalField2}</div>
  {/if}
  <div class="media-creators">
    {#each creators as creator (creator.id)}
      <div>{creator.name}</div>
    {/each}
  </div>
</div>

<style>
  .media-card {
    border: 1px solid #ccc;
    padding: 1em;
    margin: 1em;
  }

  .media-image {
    width: 100px;
  }

  .media-title {
    font-weight: bold;
    margin: 0.5em 0;
  }

  .media-creators {
    font-size: 0.9em;
    color: #666;
  }
</style>
