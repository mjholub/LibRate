<script lang="ts">
	import { onMount } from 'svelte';
	import Search from '$components/utility/Search.svelte';
	import MemberPage from '$components/member/MemberPage.svelte';
	import type { Member } from '$lib/types/member.ts';

	// we get the nickname from the last part of the URL
	export let data: { props: { nickname: string } };
	let windowWidth = 0;
	let nickname = data.props.nickname;

	if (typeof window !== 'undefined') {
		window.scrollTo(0, 0);
		windowWidth = window.innerWidth;

		// Extract the nickname from the URL
		const urlParts = window.location.pathname.split('/');
		nickname = urlParts[urlParts.length - 1];

		// Fetch user profile data when the component mounts
		onMount(() => {
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
		<MemberPage {nickname} />
	{:else}
		<p>This page is not available on mobile.</p>
	{/if}
</div>
