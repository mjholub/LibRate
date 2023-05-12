<script>
  export let items = [];

  let search = "";

  // TODO: add more versatile fuzzy finding in backend
  const searchItems = () =>
    items.filter((item) =>
      item.name.toLowerCase().includes(search.toLowerCase())
    );
</script>

<div class="search-bar">
  <input
    type="text"
    bind:value={search}
    placeholder="Search..."
    on:input={() => {}}
  />

  <ul class="search-results">
    {#each searchItems() as item (item.id)}
      <li class="search-result">{item.name}</li>
    {/each}
  </ul>
</div>
// FIXME: add graceful noscript degradation

<style>
  .search-bar {
    margin-bottom: 1em;
  }

  .search-results {
    list-style-type: none;
    padding-left: 0;
  }

  .search-result {
    margin-bottom: 0.5em;
  }
</style>
