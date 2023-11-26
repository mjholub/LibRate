<script lang="ts">
	import axios from 'axios';
	import { browser } from '$app/environment';
	import type { Member } from '$lib/types/member.ts';
	import type { NullableString } from '$lib/types/utils';

	function splitNullable(input: NullableString, separator: string): string[] {
		if (input.Valid) {
			return input.String.split(separator);
		}
		return [];
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

	const logout = async () => {
		try {
			await axios.post('/api/authenticate/logout');
			if (browser) {
				window.location.reload();
			}
		} catch (error) {
			alert(error);
		}
	};
</script>

<!-- TODO: see if this looks better as a description list -->
<div class="member-card">
	<!--  <p>Logged in as <a href="/profiles/{member.memberName}">{member.memberName}</a></p> -->
	{#if member.profilePic}
		<img class="member-image" src={member.profilePic} alt="{member.memberName}'s profile picture" />
	{:else}
		<img
			class="member-image"
			src="https://www.gravatar.com/avatar/000
    ?d=mp"
			alt="{member.memberName}'s profile picture"
		/>
	{/if}
	<div class="member-name">({member.memberName})</div>
	{#if member.bio.Valid}
		<div class="member-bio">{member.bio.String}</div>
	{/if}
	<div class="member-joined-date">Joined {regDate}</div>
	<div class="member-links">
		Other links and contact info @{member.memberName} has provided:
		{#if member.matrix.Valid}
			<p>
				<b>Matrix:</b>
				<a href="https://matrix.to/#/{matrixUser}:{matrixInstance}">{matrixUser}:{matrixInstance}</a
				>
			</p>
		{/if}
		{#if member.xmpp.Valid}
			<p><b>XMPP:</b> <a href="xmpp:{xmppUser}@{xmppInstance}">{xmppUser}@{xmppInstance}</a></p>
		{/if}
		{#if member.irc.Valid}
			<p><b>IRC:</b> <a href="irc://{ircUser}@{ircInstance}">{ircUser}@{ircInstance}</a></p>
		{/if}
		{#if member.homepage.Valid}
			<p><b>Homepage:</b> <a href={member.homepage.String}>{member.homepage}</a></p>
		{/if}
	</div>
	<button aria-label="Logout" on:click={logout} id="logout-button">Logout</button>
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
