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

<style>
  .app {
    display: flex;
    justify-content: space-between;
  }

  .left,
  .right {
    width: 30%;
  }

  .navbar {
    position: fixed;
    top: 0;
    width: 100%;
    background-color: #fff;
    padding: 10px 0;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    z-index: 1000;
  }
</style>
