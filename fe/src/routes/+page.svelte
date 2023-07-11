<script lang="ts">
	import ReviewList from '../components/review/ReviewList.svelte';
	import Auth from '../components/form/Auth.svelte';
	import Search from '../components/utility/Search.svelte';
	import MemberCard from '../components/member/MemberCard.svelte';
	import Footer from '../components/footer/footer.svelte';
	import { isAuthenticated, member as memberStore } from '../stores/members/auth.ts';
	import type { Review } from '../types/review.ts';
	import type { Member } from '../types/member.ts';

	let reviews: Review[] = [];
	let member: Member = $memberStore;
</script>

<div class="navbar">
	<Search />
</div>

<div class="app">
	<div class="left">
		<ReviewList {reviews} />
	</div>
	<div class="right">
		{#if $isAuthenticated}
			<MemberCard {member} />
		{:else}
			<Auth />
		{/if}
	</div>
	<Footer />
</div>

<style>
	.app {
		display: flex;
		justify-content: space-between;
		background-color: #333;
		color: #fff;
		padding-top: 40px;
	}

	.left,
	.right {
		padding-top: 45px;
		width: 30%;
	}

	.navbar {
		position: fixed;
		top: 0;
		width: 100%;
		background-color: #333;
		color: #fff;
		padding: 10px 0;
		box-shadow: 0 2px 4px rgba(64, 64, 64, 0.2);
		z-index: 1000;
	}

	:global(body) {
		background-color: #333;
		color: #fff;
	}
</style>
