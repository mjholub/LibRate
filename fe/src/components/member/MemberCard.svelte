<script>
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
</script>

<div class="member-card">
  <img
    class="member-image"
    src={member.image}
    alt="{member.name}'s profile picture"
  />
  <div class="member-name">{member.name}</div>
  <div class="member-bio">{member.bio}</div>
  <div class="member-joined-date">Joined {member.joinedDate}</div>
  <div class="member-links">
    Other links {member.name} has provided: // TODO: render matrix through homepage
    if provided
  </div>
  <div class="member-groups">
    {#each member.groups as group}
      <span class="member-group">{group}</span>
    {/each}
  </div>
</div>

<style>
  .member-card {
    border: 1px solid #ccc;
    padding: 1em;
    margin: 1em;
  }

  .member-image {
    width: 100px;
    height: 100px;
    border-radius: 50%;
    object-fit: cover;
    margin-bottom: 1em;
  }

  .member-name {
    font-weight: bold;
    margin-bottom: 0.5em;
  }

  .member-bio {
    font-size: 0.9em;
    color: #666;
    margin-bottom: 1em;
  }

  .member-joined-date {
    font-size: 0.8em;
    color: #999;
  }
</style>
