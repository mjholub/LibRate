<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '../../stores/members/auth.ts';
	import type { Member } from '../../types/member.ts';

	let regDate: string;
	let matrixUser: string;
	let matrixInstance: string;
	let xmppUser: string;
	let xmppInstance: string;
	let ircUser: string;
	let ircInstance: string;
	export let member: Member;
	$: {
		regDate = new Date(member.regdate).toLocaleDateString();
	}

	const splitMatrixUser = (matrixUser: string) => {
		const [user, instance] = matrixUser.split(':');
		matrixUser = user;
		matrixInstance = instance;
	};

	const splitXmppUser = (xmppUser: string) => {
		const [user, instance] = xmppUser.split('@');
		xmppUser = user;
		xmppInstance = instance;
	};

	const splitIrcUser = (ircUser: string) => {
		const [user, instance] = ircUser.split('@');
		ircUser = user;
		ircInstance = instance;
	};

	onMount(async () => {
		if (member && member.id) {
			authStore.subscribe((auth) => {
				if (auth && auth.id === member.id) {
					console.debug('member', member);
				}
			});
		}
		console.debug('member (called outside conditional): ', member);
		// split the matrix user into user and instance
		if (member.matrix) {
			splitMatrixUser(member.matrix);
		}
		if (member.xmpp) {
			splitXmppUser(member.xmpp);
		}
		if (member.irc) {
			splitIrcUser(member.irc);
		}
	});
</script>

<div class="member-card">
	<img class="member-image" src={member.profilePic} alt="{member.memberName}'s profile picture" />
	<div class="member-name">({member.memberName})</div>
	<div class="member-bio">{member.bio}</div>
	<div class="member-joined-date">Joined {regDate}</div>
	<div class="member-links">
		Other links and contact info @{member.memberName} has provided:
		{#if member.matrix}
			<p>
				<b>Matrix:</b>
				<a href="https://matrix.to/#/{matrixUser}:{matrixInstance}">{matrixUser}:{matrixInstance}</a
				>
			</p>
		{/if}
		{#if member.xmpp}
			<p><b>XMPP:</b> <a href="xmpp:{xmppUser}@{xmppInstance}">{xmppUser}@{xmppInstance}</a></p>
		{/if}
		{#if member.irc}
			<p><b>IRC:</b> <a href="irc://{ircUser}@{ircInstance}">{ircUser}@{ircInstance}</a></p>
		{/if}
		{#if member.homepage}
			<p><b>Homepage:</b> <a href={member.homepage}>{member.homepage}</a></p>
		{/if}
	</div>
	<!-- TODO: uncomment when groups are implemented
	<div class="member-groups">
		{#each member.groups as group}
			<span class="member-group">{group}</span>
		{/each}
  </div>
  -->
</div>

<style>
	.member-card {
		border: 1px solid #ccc;
		padding: 1em;
		margin: 1em;
	}

	.member-image {
		width: 100px;
		height: 100px;
		border-radius: 50%;
		object-fit: cover;
		margin-bottom: 1em;
	}

	.member-name {
		font-weight: bold;
		margin-bottom: 0.5em;
	}

	.member-bio {
		font-size: 0.9em;
		color: #666;
		margin-bottom: 1em;
	}

	.member-joined-date {
		font-size: 0.8em;
		color: #999;
	}
</style>
