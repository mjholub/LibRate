<script>
  import ReviewList from "../components/review/ReviewList.svelte";
  import Auth from "../components/form/Auth.svelte";
  import Search from "../components/utility/Search.svelte";
  import { onMount } from "svelte";

  let reviews = [];
</script>

<div class="navbar">
  <Search />
</div>

<div class="app">
  <div class="left">
    <ReviewList {reviews} />
  </div>
  <div class="right">
    <Auth />
  </div>
</div>

<style src="./App.css">
</style>
