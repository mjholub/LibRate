<script lang="ts">
	import { browser } from '$app/environment';
	import { onDestroy, onMount } from 'svelte';
	import { authStore } from '$stores/members/auth.ts';
	//import ReviewList from '../components/review/ReviewList.svelte';
	import Auth from '$components/form/Auth.svelte';
	import Search from '$components/utility/Search.svelte';
	import Footer from '$components/footer/footer.svelte';
	import MemberCard from '$components/member/MemberCard.svelte';
	import MediaCarousel from '$components/media/MediaCarousel.svelte';
	import type { Member } from '$lib/types/member.ts';
	import type { authData } from '$stores/members/auth.ts';

	let windowWidth: number;
	let isAuthenticated: boolean;
	let member: Member;
	let authstatus: authData;
	async function handleAuthentication() {
		if (browser) {
			const jwtToken = localStorage.getItem('jwtToken');
			try {
				authstatus = await authStore.authenticate(jwtToken);
				isAuthenticated = authstatus.isAuthenticated;
				console.debug('authstatus', authstatus);
			} catch (error) {
				console.error('error', error);
			}
		}
	}
	async function getMember(memberName: string) {
		member = await authStore.getMember(memberName);
	}
	if (browser) {
		onMount(async () => {
			windowWidth = window.innerWidth;
			const handleResize = () => {
				windowWidth = window.innerWidth;
			};
			window.addEventListener('resize', handleResize);
		});
	}
	onDestroy(() => {
		if (browser) {
			window.removeEventListener('resize', () => {});
		}
	});
</script>

<div class="app">
	<div class="navbar">
		<Search />
	</div>
	<div class="content">
		<div class="left">
			<MediaCarousel />
		</div>
		<div class="center">
			<div class="feed">
				<h2>Reviews feed</h2>
				<p>Coming soon...</p>
			</div>
		</div>
		<div class="right">
			{#await handleAuthentication()}
				<p>loading...</p>
			{:then}
				{#if !isAuthenticated}
					<Auth />
				{:else if isAuthenticated}
					{#await getMember(authstatus.memberName)}
						<p>loading member card...</p>
					{:then}
						<MemberCard {member} />
					{:catch}
						<p>error loading member card</p>
					{/await}
				{:else}
					<p>Client error while rendering the auth component.</p>
				{/if}
			{:catch}
				<p>error loading auth component</p>
			{/await}
		</div>
	</div>
	<div class="footer">
		<Footer />
	</div>
</div>

<style>
	:root {
		--main-bg-color: #333;
		--text-color: #fff;
		--padding-base: 20px;
	}

	.app {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
		background-color: var(--main-bg-color);
		color: var(--text-color);
	}

	.navbar {
		background-color: var(--main-bg-color);
		color: var(--text-color);
		padding: 1rem 0;
		text-align: center;
		display: block;
	}

	.content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: var(--padding-base) calc(var(--padding-base) / 2);
		flex: 1;
	}

	.left,
	.center,
	.right {
		display: flex;
		flex-direction: column;
	}

	.left {
		width: 40%;
	}

	.center {
		width: 30%;
		justify-content: center;
	}

	.feed {
		text-align: center;
	}

	.right {
		width: 30%;
	}

	.footer {
		padding: var(--padding-base);
		background-color: var(--main-bg-color);
		color: var(--text-color);
		text-align: center;
	}
</style>
