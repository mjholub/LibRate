<script>
  import { onMount } from "svelte";

  export let items = [];

  let search = "";

  // function to fetch data from the backend based on the search term
  async function searchItems() {
    const response = await fetch("http://localhost:3000/api/search", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ search }),
    });

    const data = await response.json();

    // If the backend responds with the filtered items
    // update items with the new data
    items = data;
  }

  onMount(() => {
    searchItems();
  });
</script>

<div class="search-bar">
  <input
    type="text"
    bind:value={search}
    placeholder="Enter search keywords..."
    on:input={searchItems}
  />
  <button on:click={searchItems}>Search</button>
</div>

<div>
  <ul class="search-results">
    {#each searchItems() as item (item.id)}
      <li class="search-result">{item.name}</li>
    {/each}
  </ul>
</div>

<style>
  .search-bar {
    margin-bottom: 0.8em;
  }

  .search-results {
    list-style-type: none;
    padding-left: 0;
  }

  .search-result {
    margin-bottom: 0.5em;
  }
</style>
