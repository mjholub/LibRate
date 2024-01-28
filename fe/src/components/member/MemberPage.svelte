<script lang="ts">
	import { memberStore } from '$stores/members/getInfo';
	import MemberCard from './MemberCard.svelte';
	import type { Review } from '$lib/types/review';
	import type { Member } from '$lib/types/member';

	export let nickname: string;
	let canView: boolean = false;
	let member: Member;
	console.info('fetching member info for', nickname);

	const jwtToken = localStorage.getItem('jwtToken');
	const getMember = async (nickname: string) => {
		try {
			if (jwtToken) {
				member = await memberStore.getMember(jwtToken, nickname);
			} else {
				// this can still work for public profiles
				member = await memberStore.getMember('', nickname);
			}
			if (member) {
				canView = true;
			}
		} catch (error) {
			canView = false;
		}
	};
</script>

<div class="member-page">
	<div class="member-page-content">
		{#await getMember(nickname)}
			<p>loading...</p>
		{:then}
			{#if canView}
				<div class="member-info">
					<MemberCard {member} />
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
</style>
