<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { authStore } from '../stores/members/auth.ts';
	//import ReviewList from '../components/review/ReviewList.svelte';
	import Auth from '../components/form/Auth.svelte';
	import Search from '../components/utility/Search.svelte';
	import MemberCard from '../components/member/MemberCard.svelte';
	import Footer from '../components/footer/footer.svelte';
	import MediaCarousel from '../components/media/MediaCarousel.svelte';
	import type { Review } from '../types/review.ts';
	import type { Member } from '../types/member.ts';
	import type { UUID } from '../types/utils.ts';
  import type { AuthStoreState } from '../stores/members/auth.ts';

	let windowWidth: number;
  let authState: AuthStoreState = $authStore.state;
	if (browser) {
		onMount(() => {
			windowWidth = window.innerWidth;
			const handleResize = () => {
				windowWidth = window.innerWidth;
			};
			const handleAuth = async () => {
				await authStore.authenticate();
			};
			handleAuth();
			window.addEventListener('resize', handleResize);

			return () => {
				window.removeEventListener('resize', handleResize);
			};
		});
	}

	//let reviews: Review[] = [];
	let member: Member = $authStore.member;
</script>

<div class="navbar">
	<Search />
</div>

<div class="app">
	<div class="left" class:hidden={windowWidth <= 768}>
		<MediaCarousel />
	</div>
	<div class="right">
		{#if $authStore.isAuthenticated}
			<MemberCard {member} />
		{:else}
			<Auth />
		{/if}
	</div>
</div>
<div class="footer">
	<Footer />
</div>

<style>
	.hidden {
		display: none;
	}
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
		width: 35%;
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

	.footer {
		position: bottom;
		bottom: 0;
		width: 100%;
		align-items: center;
	}

	:global(body) {
		background-color: #333;
		color: #fff;
	}
</style>
