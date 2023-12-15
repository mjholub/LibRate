<script lang="ts">
	import { memberStore } from '$stores/members/getInfo';
	import MemberCard from './MemberCard.svelte';
	import ReviewList from '../review/ReviewList.svelte';
	import type { Review } from '$lib/types/review';
	import type { Member } from '$lib/types/member';

	export let nickname: string;
	let member: Member;
	console.info('fetching member info for', nickname);

	const getMember = async (nickname: string) => {
		const jwtToken = localStorage.getItem('jwtToken');
		if (jwtToken === null) {
			console.error('jwtToken is null');
			return;
		}
		member = await memberStore.getMember(jwtToken, nickname);
	};

	let reviews: Review[];
</script>

<div class="member-page">
	<div class="member-page-content">
		{#await getMember(nickname)}
			<p>loading...</p>
		{:then}
			<div class="member-info">
				<MemberCard {member} />
			</div>
		{:catch error}
			<p>error: {error.message}</p>
		{/await}
		<!--
		<div class="reviews">
			<ReviewList {reviews} />
		</div>
        -->
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
