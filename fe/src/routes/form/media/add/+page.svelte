<script lang="ts">
	import type { authData } from '$stores/members/auth.ts';
	import Auth from '$components/form/Auth.svelte';
	import AddMedia from '$components/form/AddMedia.svelte';
	import Footer from '$components/footer/footer.svelte';
	import Header from '$components/header/Header.svelte';

	import { browser } from '$app/environment';
	let authstatus: authData;
	let isAuthenticated: boolean;
	import { authStore } from '$stores/members/auth.ts';

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
</script>

{#await handleAuthentication()}
	<span class="spinner" />
{:then}
	{#if !isAuthenticated}
		<Header authenticated={isAuthenticated} nickname="" />
		<p>Log in first</p>
		<Auth />
		<Footer />
	{:else}
		<Header authenticated={isAuthenticated} nickname={authstatus.memberName} />
		<AddMedia nickname={authstatus.memberName} />
		<Footer />
	{/if}
{:catch error}
	<p>{error.message}</p>
{/await}

<style>
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
