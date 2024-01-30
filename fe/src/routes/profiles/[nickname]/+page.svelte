<script lang="ts">
	import { onMount } from 'svelte';
	import { _ } from 'svelte-i18n';
	import { authStore } from '$stores/members/auth.ts';
	import type { authData } from '$stores/members/auth.ts';
	import { filterXSS } from 'xss';
	import Header from '$components/header/Header.svelte';
	import MemberPage from '$components/member/MemberPage.svelte';

	let windowWidth = 0;
	// must initialize zeroed authData for header to work properly instead of loading infinitely
	let authstatus: authData = { isAuthenticated: false, memberName: '' };
	let params: { slug?: string; page?: string } = {};
	console.debug('params', params.toString());
	$: nickname = params?.slug ?? '';
	const jwtToken = localStorage.getItem('jwtToken') ?? '';

	const handleAuthentication = async () => {
		try {
			authstatus = await authStore.authenticate(jwtToken);
		} catch (error) {
			console.error(error);
		}
	};

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
	{#await handleAuthentication()}
		<p>{$_('loading')} {$_('header')}</p>
		<span class="dot-flashing" />
	{:then}
		<Header authenticated={authstatus.isAuthenticated} nickname={authstatus.memberName} />
	{:catch error}
		<p>{$_('error')} {$_('header')}: {error.message}</p>
	{/await}
</div>
{#if nickname}
	<h1 class="title">{$_('profile_of')} {nickname}</h1>
	<div class="profile">
		<MemberPage {nickname} />
	</div>
{:else}
	<p>Member not found</p>
{/if}
