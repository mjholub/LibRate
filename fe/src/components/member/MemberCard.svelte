<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { authStore, initialAuthState } from '$stores/members/auth.ts';
	import type { Member } from '$lib/types/member.ts';

	function splitNullable(input: string | null, separator: string): string[] {
		return input ? input.split(separator) : [];
	}

	let matrixInstance: string,
		matrixUser: string,
		xmppUser: string,
		xmppInstance: string,
		ircUser: string,
		ircInstance: string;

	$: {
		matrixInstance = splitNullable(member.matrix, ':')[1];
		matrixUser = splitNullable(member.matrix, ':')[0];
		xmppUser = splitNullable(member.xmpp, '@')[0];
		xmppInstance = splitNullable(member.xmpp, '@')[1];
		ircUser = splitNullable(member.irc, '@')[0];
		ircInstance = splitNullable(member.irc, '@')[1];
	}

	let regDate: string;
	export let member: Member;
	$: {
		regDate = new Date(member.regdate).toLocaleDateString();
	}

	onMount(async () => {
		if (member && member.id) {
			authStore.subscribe((auth) => {
				if (auth && auth.id === member.id) {
					console.debug('member', member);
				}
			});
		}
		console.debug('member (called outside conditional): ', member);
	});

	onDestroy(() => {
		authStore.set(initialAuthState);
	});
</script>

<!-- TODO: see if this looks better as a description list -->
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

<!-- TODO: use CSS variables -->
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
