<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$stores/members/auth';
	import getMemberProps from './MemberCard.svelte';
	import MemberCard from './MemberCard.svelte';
	import ReviewList from '../review/ReviewList.svelte';
	import type { Review } from '$lib/types/review';
	import type { Member } from '$lib/types/member';

	export let nickname: string;
	console.info('fetching member info for', nickname);
	let member: Member;

	onMount(async () => {
		member = await authStore.getMember(nickname);
		if (member && member.uuid) {
			new getMemberProps({ target: document.body, props: { member } });
		}
	});

	// TODO: fetch this data based on the member's id
	let reviews: Review[];
</script>

<div class="member-page">
	<div class="member-page-content">
		<div class="member-info">
			<MemberCard {member} />
		</div>

		<div class="reviews">
			<ReviewList {reviews} />
		</div>
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
