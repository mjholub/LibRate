<script lang="ts">
	import { onMount } from 'svelte';
	import { memberStore } from '$stores/members/getInfo';
	import MemberCard from './MemberCard.svelte';
	import type { Review } from '$lib/types/review';
	import type { Member } from '$lib/types/member';

	export let nickname: string;
	export let viewerName: string;
	let canView: boolean = false;
	let member: Member;
	console.info('fetching member info for', nickname);

	const jwtToken = localStorage.getItem('jwtToken');
	const getMember = async (nickname: string) => {
		if (jwtToken === null) {
			console.error('jwtToken is null');
			return;
		}
		member = await memberStore.getMember(jwtToken, nickname);
	};

	onMount(async () => {
		if (jwtToken === null) {
			// we use "" since public accounts do not require a jwtToken
			canView = await memberStore.verifyViewablity('', viewerName, nickname);
		} else {
			canView = await memberStore.verifyViewablity(jwtToken, viewerName, nickname);
		}
	});
</script>

<div class="member-page">
	<div class="member-page-content">
		{#await getMember(nickname)}
			<p>loading...</p>
		{:then}
			{#if canView}
				<div class="member-info">
					<MemberCard {member} showLogout={false} />
				</div>
			{:else}
				<p>Account not found or private.</p>
			{/if}
		{:catch error}
			<p>error: {error.message}</p>
		{/await}
	</div>
</div>

<style>
	.member-page {
		padding: 1em;
	}

	.member-page-content {
		display: flex;
		gap: 2em;
	}

	.member-info {
		flex: 1 0 40%;
	}

	.reviews {
		flex: 1 0 60%;
	}
</style>
