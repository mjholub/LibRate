<script lang="ts">
	import { browser } from '$app/environment';
	import { authStore } from '$stores/members/auth.ts';
	import { _ } from 'svelte-i18n';
	import ErrorModal from '$components/modal/ErrorModal.svelte';
	import Auth from '$components/form/Auth.svelte';
	import Header from '$components/header/Header.svelte';
	import Footer from '$components/footer/footer.svelte';
	import MemberCard from '$components/member/MemberCard.svelte';
	import MediaCarousel from '$components/media/MediaCarousel.svelte';
	import type { Member } from '$lib/types/member.ts';
	import type { Review } from '$lib/types/review.ts';
	import type { authData } from '$stores/members/auth.ts';
	import { memberStore, memberInfo } from '$stores/members/getInfo';
	import type { CustomHttpError } from '$lib/types/error';
	import ReviewCard from '$components/review/ReviewCard.svelte';

	let windowWidth: number;
	let isAuthenticated: boolean;
	let member: Member;
	let authstatus: authData;
	let errors: CustomHttpError[];
	let reviews: Review[] = [];
	$: errors = [];
	$: member = memberInfo;
	async function handleAuthentication() {
		if (browser) {
			const jwtToken = localStorage.getItem('jwtToken');
			try {
				if (jwtToken === null) {
					errors.push({
						message: 'Missing JWT token',
						status: 401
					});
					errors = [...errors];
					return;
				}
				authstatus = await authStore.authenticate(jwtToken);
				isAuthenticated = authstatus.isAuthenticated;
				console.debug('authstatus', authstatus);
			} catch (error) {
				errors.push({
					message: error as string,
					status: 500
				});
				errors = [...errors];
			}
		}
	}
	async function getMember(memberName: string) {
		const jwtToken = localStorage.getItem('jwtToken');
		if (jwtToken === null) {
			errors.push({
				message: 'Missing JWT token',
				status: 401
			});
			errors = [...errors];
			return;
		}
		try {
			member = await memberStore.getMember(jwtToken, memberName);
		} catch (error) {
			errors.push({
				message: error as string,
				status: 500
			});
			errors = [...errors];
		}
	};
</script>

<div class="app">
	<div class="navbar">
		{#await handleAuthentication()}
			<p>Loading header</p>
			<span class="spinner" />
		{:then}
			{#if !isAuthenticated}
				<Header authenticated={isAuthenticated} nickname="" />
			{:else}
				<Header authenticated={isAuthenticated} nickname={authstatus.memberName} />
			{/if}
		{:catch error}
			<ErrorModal
				errorMessages={[
					{
						message: "Couldn't load header",
						status: error.status
					}
				]}
			/>
		{/await}
	</div>
	<div class="content">
		<div class="left">
			<MediaCarousel authenticated={isAuthenticated} />
		</div>
		<div class="center">
			<div class="feed">
				<h2>{$_('reviews_feed')}</h2>
				{#if reviews.length > 0}
					{#each reviews as review}
						<ReviewCard {review} />
					{/each}
				{:else}
					<p>{$_('no_reviews_found')}</p>
				{/if}
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
						<span class="spinner" />
					{:then}
						<MemberCard {member} />
					{:catch}
						<ErrorModal errorMessages={errors} />
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
	
	@media (max-width: 768px) {
		.left {
			display: none !important;
		}
		.right {
			max-width: 45% !important;
		}
		.center {
			max-width: 50% !important;
		}
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
		padding: 0.3rem 0.1rem 0.6rem 0.1rem;
		text-align: left;
		display: unset;
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
		max-width: 34%;
	}

	.center {
		max-width: 33%;
		justify-content: center;
	}

	.feed {
		text-align: center;
	}

	.right {
		max-width: 33%;
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
