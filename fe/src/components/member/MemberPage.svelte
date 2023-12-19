<script lang="ts">
	import { onMount } from 'svelte';
	import { memberInfo, memberStore } from '$stores/members/getInfo';
	import { authStore } from '$stores/members/auth';
	import MemberCard from './MemberCard.svelte';
	import ReviewList from '../review/ReviewList.svelte';
	import type { Review } from '$lib/types/review';
	import type { Member } from '$lib/types/member';

	export let nickname: string;
	let member: Member;
	const jwtToken = localStorage.getItem('jwtToken') || '';

	const getMember = async (nickname: string) => {
		const jwtToken = localStorage.getItem('jwtToken');
		if (jwtToken === null) {
			console.error('jwtToken is null');
			return;
		}
		member = await memberStore.getMember(jwtToken, nickname);
	};

	const displayPrivacyInfo = (rule: string) => {
		const sel = document.getElementById(rule);
		if (sel) {
			sel.style.display = 'block';
		}
	};

	const checkProfilePrivacy = async () => {
		switch (member.visibility) {
			case 'private':
				displayPrivacyInfo('private-or-nonexistent');
			case 'followers_only':
				const authData = await authStore.authenticate(jwtToken);
				if (authData.isAuthenticated) {
					// 1st arg is follower, 2nd is the followee
					const isFollower = await memberStore.checkFollowing(
						jwtToken,
						authData.memberName,
						nickname
					);
					isFollower ? displayPrivacyInfo('public') : displayPrivacyInfo('followers_only');
				}
			case 'public':
				displayPrivacyInfo('public');
			// TODO: add local (instance)-only accounts
		}
	};

	onMount(() => {
		getMember(nickname);
		checkProfilePrivacy();
	});

	let reviews: Review[];
</script>

<div class="member-page">
	<div class="member-page-content">
		{#await getMember(nickname)}
			<p>loading...</p>
		{:then}
			<div class="member-info" id="public">
				<MemberCard {member} />
			</div>
			<div id="private-or-nonexistent">
				<p>
					@{member.memberName} is a private account, you cannot interact with them even if signed in
				</p>
			</div>
			<div id="followers_only">
				<p>
					@{member.memberName} is a followers-only account, you cannot interact with them unless you
					follow them (and they accept your follow request if they have that setting enabled)
				</p>
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
