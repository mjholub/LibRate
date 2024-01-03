<script lang="ts">
	import { onMount } from 'svelte';
	import { filterXSS } from 'xss';
	import Search from '$components/utility/Search.svelte';
	import MemberPage from '$components/member/MemberPage.svelte';

	let windowWidth = 0;
	let params: { slug?: string; page?: string } = {};
	console.debug('params', params.toString());
	$: nickname = params?.slug ?? '';

	if (typeof window !== 'undefined') {
		window.scrollTo(0, 0);
		windowWidth = window.innerWidth;

		// Fetch user profile data when the component mounts
		onMount(() => {
			// Extract the nickname from the URL
			const urlParts = window.location.pathname.split('/');
			nickname = filterXSS(urlParts[urlParts.length - 1]);

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
{#if nickname}
	<h1 class="title">Profile of {nickname}</h1>
	<div class="profile">
		<MemberPage {nickname} />
	</div>
{:else}
	<p>Member not found</p>
{/if}
