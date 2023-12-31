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
	import { memberStore } from '$stores/members/getInfo';

	let windowWidth: number;
	let isAuthenticated: boolean;
	let member: Member;
	let authstatus: authData;
	async function handleAuthentication() {
		if (browser) {
			const jwtToken = localStorage.getItem('jwtToken');
			try {
				if (jwtToken === null) {
					console.error('jwtToken is null');
					return;
				}
				authstatus = await authStore.authenticate(jwtToken);
				isAuthenticated = authstatus.isAuthenticated;
				console.debug('authstatus', authstatus);
			} catch (error) {
				console.error('error', error);
			}
		}
	}
	async function getMember(memberName: string) {
		const jwtToken = localStorage.getItem('jwtToken');
		if (jwtToken === null) {
			console.error('jwtToken is null');
			return;
		}
		member = await memberStore.getMember(jwtToken, memberName);
	}
	if (browser) {
		onMount(async () => {
			windowWidth = window.innerWidth;
			const handleResize = () => {
				windowWidth = window.innerWidth;
				const left = document.getElementById('left');
				if (left) {
					// if the window size is less than 768px, hide the left column
					if (windowWidth < 768) {
						left.style.display = 'none';
					} else {
						left.style.display = 'block';
					}
				}
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
		<div id="left">
			<MediaCarousel authenticated={isAuthenticated} />
		</div>
		<div class="center">
			<div class="feed">
				<h2>Reviews feed</h2>
				<p>Coming soon...</p>
			</div>
		</div>
		<div class="right">
			{#await handleAuthentication()}
				<span class="spinner" />
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
		text-align: left;
		display: block;
	}

	.content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: var(--padding-base) calc(var(--padding-base) / 2);
		flex: 1;
	}

	div#left,
	.center,
	.right {
		display: flex;
		flex-direction: column;
	}

	div#left {
		width: 34%;
	}

	.center {
		width: 33%;
		justify-content: center;
	}

	.feed {
		text-align: center;
	}

	.right {
		width: 33%;
	}

	.footer {
		padding: var(--padding-base);
		background-color: var(--main-bg-color);
		color: var(--text-color);
		text-align: center;
	}

	.spinner {
		border: 4px solid rgba(0, 0, 0, 0.1);
		border-top: 4px solid #3498db;
		border-radius: 50%;
		width: 20px;
		height: 20px;
		animation: spin 1s linear infinite;
		margin-left: 10px;
		display: inline-block;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>
