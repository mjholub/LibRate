<script lang="ts">
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import Search from '$components/utility/Search.svelte';
	import MemberPage from '$components/member/MemberPage.svelte';
	import type { Member } from '$types/member.ts';

	// we get the nickname from the last part of the URL
	let nickname = '';
	const userProfile = writable<Member | null>(null);
	let windowWidth = 0;

	function fetchUserProfile(nickname: string) {
		console.log('Fetching user profile for:', nickname);
		fetch(`/api/members/${nickname}/info`)
			.then((response) => response.json())
			.then((data) => {
				if (data.error) {
					// Handle the case where the user is not found
					userProfile.set(null); // Set the store to null in case of an error
				} else {
					// Set the retrieved user profile data to the store
					userProfile.set(data.data);
				}
			})
			.catch((error) => {
				console.error('Error fetching user profile:', error);
				userProfile.set(null); // Set the store to null in case of an error
			});
	}

	if (typeof window !== 'undefined') {
		window.scrollTo(0, 0);
		windowWidth = window.innerWidth;

		// Extract the nickname from the URL
		const urlParts = window.location.pathname.split('/');
		nickname = urlParts[urlParts.length - 1];

		// Fetch user profile data when the component mounts
		onMount(() => {
			fetchUserProfile(nickname);

			const handleResize = () => {
				windowWidth = window.innerWidth;
			};
			window.addEventListener('resize', handleResize);

			return () => {
				window.removeEventListener('resize', handleResize);
			};
		});
	}
</script>

<div class="navbar">
	<Search />
</div>
<h1 class="title">Profile of {nickname}</h1>
<div class="profile">
	{#if windowWidth > 768}
		{#if $userProfile}
			<MemberPage {nickname} />
		{:else}
			<p>Profile not found or an error occured.</p>
		{/if}
	{:else}
		<p>This page is not available on mobile.</p>
	{/if}
</div>
