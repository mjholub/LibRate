<script>
  import MemberCard from "./MemberCard.svelte";
  import ReviewList from "./ReviewList.svelte";

  export let member = {
    id: "",
    name: "",
    fullName: "",
    image: "",
    bio: "",
    matrix: "",
    xmpp: "",
    irc: "",
    microblog: "",
    homepage: "",
    bookwyrm: "",
    funkwhale: "",
    joinedDate: "",
    isCreator: false,
    isActive: false,
    isAdmin: false,
    groups: [],
  };

  // TODO: fetch this data based on the member's id
  let reviews = [];
</script>

<div class="member-page">
  <div class="member-page-content">
    <div class="member-info">
      <MemberCard {member} />
    </div>

    <div class="reviews">
      <ReviewList {reviews} />
    </div>
  </div>
</div>

<style>
  .member-page {
    padding: 1em;
  }

  .member-page-content {
    display: flex;
    gap: 2em;
  }

  .member-info {
    flex: 1 0 40%;
  }

  .reviews {
    flex: 1 0 60%;
  }
</style>
