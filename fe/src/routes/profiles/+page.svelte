<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import Search from '../../components/utility/Search.svelte';
	import MemberPage from '../../components/member/MemberPage.svelte';
	let windowWidth: number;
	let nickname: string;
	if (browser) {
		nickname = window.location.pathname.split('/')[2];
	} else {
		nickname = '';
	}
	if (browser) {
		onMount(() => {
			// fetch member data based on the nick, which is the last part of the url
			// e.g. /profiles/+page -> page
			if (typeof window !== 'undefined') {
				window.scrollTo(0, 0);
				windowWidth = window.innerWidth;
				const handleResize = () => {
					windowWidth = window.innerWidth;
				};
				window.addEventListener('resize', handleResize);

				return () => {
					window.removeEventListener('resize', handleResize);
				};
			} else {
				windowWidth = 0;
			}
		});
	}
</script>

<div class="navbar">
	<Search />
</div>
<div class="profile">
	{#if windowWidth > 768}
		<MemberPage {nickname} />
	{:else}
		<p>This page is not available on mobile.</p>
	{/if}
</div>
