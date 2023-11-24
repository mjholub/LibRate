<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import { authStore } from '$stores/members/auth.ts';
	//import ReviewList from '../components/review/ReviewList.svelte';
	import Auth from '$components/form/Auth.svelte';
	import Search from '$components/utility/Search.svelte';
	// FIXME: member card not rendering upon login
	//import MemberCard from '$components/member/MemberCard.svelte';
	import Footer from '$components/footer/footer.svelte';
	import MemberCard from '$components/member/MemberCard.svelte';
	import MediaCarousel from '$components/media/MediaCarousel.svelte';
	import type { Member } from '$lib/types/member.ts';
	import type { AuthStoreState } from '$stores/members/auth.ts';

	let windowWidth: number;
	$: authState = $authStore;
	let authenticating = true;
	let member: Member;
	async function handleAuthentication() {
		let unsubscribe: () => void;

		function updateAuthState(newAuthState: AuthStoreState) {
			if (newAuthState.isAuthenticated) {
				console.debug('User is authenticated', newAuthState);
				unsubscribe(); // Unsubscribe to avoid further updates
			}
		}

		unsubscribe = authStore.subscribe(updateAuthState);
		await authStore.authenticate();
	}
	if (browser) {
		onMount(() => {
			windowWidth = window.innerWidth;
			const handleResize = () => {
				windowWidth = window.innerWidth;
			};
			window.addEventListener('resize', handleResize);

			const wasAuthConfirmationDisplayed = JSON.parse(
				localStorage.getItem('wasAuthConfirmationDisplayed') || 'false'
			);
			if ($authStore.isAuthenticated && !wasAuthConfirmationDisplayed) {
				alert('Logged in');
				localStorage.setItem('wasAuthConfirmationDisplayed', JSON.stringify(true));
			}

			handleAuthentication();

			authStore.subscribe(async (newAuthState) => {
				authState = newAuthState;
				const sessionCookie = document.cookie.includes('session=');
				if (sessionCookie) {
					// using try-cacth to avoid unhandled promise rejection
					try {
						const res = await fetch(`/api/authenticate`);
						res.ok ? (authState.isAuthenticated = true) : (authState.isAuthenticated = false);
					} catch (err) {
						console.error(err);
					}
				}
				const localStorageData = localStorage.getItem('member');
				if (localStorageData) {
					const parsedData = JSON.parse(localStorageData);

					// Extract the relevant properties from the "data" object
					const data = parsedData?.data || {}; // Ensure "data" exists

					// Create a new Member object with the correct types
					member = {
						id: data.id || 0, // Provide a default value if id is missing
						memberName: data.memberName || '',
						displayName: data.displayname?.String || null,
						email: data.email || '',
						profilePic: data.profilePic || null,
						bio: data.bio?.String || null,
						matrix: data.matrix?.String || null,
						xmpp: data.xmpp?.String || null,
						irc: data.irc?.String || null,
						homepage: data.homepage?.String || null,
						regdate: new Date(data.regdate) || null,
						roles: data.roles || [],
						visibility: data.visibility || 'private'
					};
				} else {
					member = {
						id: 0,
						memberName: '',
						displayName: null,
						email: '',
						profilePic: null,
						bio: null,
						matrix: null,
						xmpp: null,
						irc: null,
						homepage: null,
						regdate: new Date(),
						roles: [],
						visibility: 'private'
					};
					console.warn('No member data found in local storage');
				}
				authenticating = false;
			});

			return () => {
				window.removeEventListener('resize', handleResize);
			};
		});
	}

	//let reviews: Review[] = [];
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
			{#if authenticating || !authState.isAuthenticated}
				<Auth />
			{:else if authState.isAuthenticated}
				{#if member}
					<p>Logged in as <a href="/profiles/{member.memberName}">{member.memberName}</a></p>
					<MemberCard {member} />
				{/if}
			{:else}
				<p>Client error while rendering the auth component.</p>
			{/if}
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
