<script lang="ts">
	import axios from 'axios';
	import { XIcon, MaximizeIcon, EditIcon } from 'svelte-feather-icons';
	import { onMount, onDestroy } from 'svelte';
	import type { Member } from '$lib/types/member.ts';
	import type { NullableString } from '$lib/types/utils';
	import { browser } from '$app/environment';
	import { authStore } from '$stores/members/auth';
	import UpdateBio from '$components/form/UpdateBio.svelte';
	import { openFilePicker, getMaxFileSize } from '$stores/form/upload';
	import type { CustomHttpError } from '$lib/types/error';
	import ErrorModal from '$components/modal/ErrorModal.svelte';

	const tooltipMessage = 'Change profile picture (max. 400x400px)';
	export let member: Member;
	export let showLogout: boolean = true;
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
		regDate = new Date(member.regdate).toLocaleDateString();
	}

	let regDate: string;
	let maxFileSize: number;
	let maxFileSizeString: string;
	let errorMessages: CustomHttpError[] = [];
	onMount(async () => {
		maxFileSize = await getMaxFileSize();
	});

	onDestroy(() => {
		maxFileSize = 0;
	});

	let uploaded: boolean;

	$: {
		maxFileSizeString = `${(maxFileSize / 1024 / 1024).toFixed(2)} MB`;
		uploaded = false;
	}

	const logout = async () => {
		try {
			const csrfToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('csrf_'))
				?.split('=')[1];
			if (csrfToken) {
				authStore.logout(csrfToken);
				authStore.set({ isAuthenticated: false });
			}
			if (browser) {
				window.location.reload();
				localStorage.removeItem('jwtToken');
			}
		} catch (error) {
			errorMessages.push({
				message: 'Error logging out: ' + error,
				status: 500
			});
			errorMessages = [...errorMessages];
			console.error(error);
		}
	};

	let isUploading = false;
	let showModal = false;

	const toggleModal = () => {
		showModal = !showModal;
	};
	const jwtToken = localStorage.getItem('jwtToken');

	const handleFileSelection = async (e: Event) => {
		return new Promise(async (resolve, reject) => {
			if (browser) {
				const fileInput = e.target as HTMLInputElement;
				const file = fileInput.files?.[0];
				if (!file) {
					return;
				}
				// checking file size is done in the backend
				isUploading = true;
				const formData = new FormData();
				formData.append('fileData', file);
				formData.append('imageType', 'profile');
				formData.append('member', member.memberName);
				let csrfToken: string | undefined;
				csrfToken = document.cookie
					.split('; ')
					.find((row) => row.startsWith('csrf_'))
					?.split('=')[1];
				const response = await axios.post('/api/upload/image', formData, {
					headers: {
						'Content-Type': 'multipart/form-data',
						Authorization: `Bearer ${jwtToken}`,
						'X-CSRF-Token': csrfToken || ''
					}
				});
				if (response.status !== 201) {
					errorMessages.push({
						message: 'Error uploading profile picture',
						status: response.status
					});
					errorMessages = [...errorMessages];
					reject(response.status);
				}
				console.log('uploaded!');
				uploaded = true;
				isUploading = false;
				// console.log(response.data);
				const pic_id = response.data.data.pic_id;
				console.log(pic_id);
				const confirmSave = confirm('Save new profile picture?');
				if (confirmSave) {
					const res = await axios.patch(
						`/api/members/update/${member.memberName}?profile_pic_id=${pic_id}`,
						{
							memberName: member.memberName
						},
						{
							headers: {
								'Content-Type': 'multipart/form-data',
								Authorization: `Bearer ${jwtToken}`,
								'X-CSRF-Token': csrfToken || ''
							}
						}
					);
					if (res.status !== 200) {
						errorMessages.push({
							message: 'Error updating profile picture',
							status: res.status
						});
						errorMessages = [...errorMessages];
						reject();
					}
				} else {
					const res = await axios.delete(`/api/upload/image/${pic_id}`, {
						headers: {
							Authorization: `Bearer ${jwtToken}`,
							'X-CSRF-Token': csrfToken || ''
						}
					});
					if (res.status !== 200) {
						errorMessages.push({
							message: 'Error deleting profile picture',
							status: res.status
						});
						errorMessages = [...errorMessages];
						reject();
					}
				}
				isUploading = false;
				resolve(void 0);
			}
		});
	};

	$: {
		if (uploaded) {
			member.profile_pic = `/static/img/profile/${member.memberName}.png`;
		}
	}

	const reloadBio = (event: CustomEvent) => {
		member.bio.Valid = true;
		member.bio.String = event.detail.newBio;
	};
</script>

<div class="member-card">
	{#if member.profile_pic}
		<div class="member-image-container">
			<img
				class="member-image"
				src={member.profile_pic}
				alt="{member.memberName}'s profile picture"
			/>
			<button
				aria-label="View full image"
				on:click={toggleModal}
				on:keypress={toggleModal}
				id="expand-image-button"
			>
				<span class="tooltip" aria-label="View full image" />
				<div class="maximize-button">
					<MaximizeIcon class="maximize-icon" />
				</div>
			</button>
			<button
				aria-label="Change profile picture (max. {maxFileSizeString})"
				id="change-profile-pic-button"
				on:click={() => openFilePicker(handleFileSelection, 'image/*')}
				on:keypress={() => openFilePicker(handleFileSelection, 'image/*')}
				><span class="tooltip" aria-label={tooltipMessage} />
				<div class="edit-button">
					<EditIcon />
				</div>
			</button>
			{#if isUploading}
				<div class="spinner" />
			{/if}
			{#if errorMessages.length > 0}
				<ErrorModal {errorMessages} />
			{/if}
		</div>
	{:else}
		<div class="member-image-container">
			<img
				class="member-image"
				src="/static/avatar-placeholder.png"
				alt="{member.memberName}'s profile picture"
			/>
			<button
				aria-label="Change profile picture (max. {maxFileSizeString})"
				id="change-profile-pic-button"
				on:click={() => openFilePicker(handleFileSelection, 'image/*')}
				on:keypress={() => openFilePicker(handleFileSelection, 'image/*')}
				><span class="tooltip" aria-label={tooltipMessage} />
				<div class="edit-button">
					<EditIcon />
				</div>
			</button>
			{#if isUploading}
				<div class="spinner" />
			{/if}
		</div>
	{/if}
	<div class="member-name">@{member.memberName}</div>
	{#if member.bio.Valid}
		<div id="member-bio">{member.bio.String}</div>
		<UpdateBio
			memberName={member.memberName}
			isBioPresent={member.bio.Valid}
			bio={member.bio.String}
			on:bioUpdated={reloadBio}
		/>
	{:else}
		<UpdateBio
			memberName={member.memberName}
			isBioPresent={member.bio.Valid}
			bio=""
			on:bioUpdated={reloadBio}
		/>
		/>
	{/if}
	<div class="member-joined-date">Joined {regDate}</div>
	Other links and contact info for @{member.memberName}:
	<!-- TODO: replace with user-defined custom fields, like on e.g. pleroma -->
	{#if member.matrix.Valid}
		<p>
			<b>Matrix:</b>
			<a href="https://matrix.to/#/{matrixUser}:{matrixInstance}">{matrixUser}:{matrixInstance}</a>
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
{#if showLogout}
	<button aria-label="Logout" on:click={logout} id="logout-button">Logout</button>
{/if}
{#if showModal}
	<div class="modal">
		<img src={member.profile_pic} alt="{member.memberName}'s profile picture" />
		<div class="close-button">
			<XIcon />
		</div>
	</div>
{/if}

<style>
	:root {
		--member-card-border-radius: 0.25em;
		--member-card-background-color: #1f1f1f;
		--logout-button-align: right;
		--logout-button-padding-top: 0.25em;
		--close-button-align: right;
		--close-button-width: 1.2em;
		--close-button-height: 1.2em;
		--icon-color: #ffcbcc;
		--button-bg: #60605190;
		--button-radius: 20%;
		--change-profile-pic-btn-right: -60%;
		--change-profile-pic-btn-top: -2.5rem;
		--expand-image-btn-right: 1.15rem;
		--expand-image-btn-bottom: 0.65rem;
	}

	:xicon {
		width: 1.2em;
		height: 1.2em;
		fill: none;
	}

	#change-profile-pic-button {
		position: relative !important;
		right: var(--change-profile-pic-btn-right) !important;
		top: var(--change-profile-pic-btn-top) !important;
		margin-right: 0.7em !important;
		width: 1.75em !important;
		height: 1.75em !important;
		padding: 0 0 0.25em 0;
		border-radius: var(--button-radius);
		background: var(--button-bg);
	}

	.edit-button {
		width: 1em;
		height: 1em;
		fill: none;
		display: contents !important;
		/* calculate contrast between component background and element background */
		mix-blend-mode: difference;
		color: var(--icon-color);
		position: relative;
	}

	.edit-button:hover {
		fill: #fafafa;
	}

	.member-image {
		width: 5em;
		height: 5em;
		border-radius: 50%;
		object-fit: cover;
		margin-bottom: 1em;
		max-width: 100%;
		font-size: smaller;
	}

	.member-image-container {
		position: relative;
		display: inline-block;
	}

	.member-card {
		border: 1px solid #ccc;
		padding: 1em;
		margin: 1em;
		border-radius: var(--member-card-border-radius);
		background-color: var(--member-card-background-color);
	}

	button#expand-image-button {
		background: var(--button-bg);
		border-radius: var(--button-radius);
		width: 1.75em !important;
		height: 1.75em !important;
		position: relative;
		bottom: var(--expand-image-btn-bottom) !important;
		right: var(--expand-image-btn-right) !important;
		padding-right: 0.1em;
	}

	.maximize-button {
		display: inherit !important;
		width: 1em;
		height: 1em;
		fill: none;
		color: var(--icon-color);
		margin-top: -0.35em !important;
		padding: 0 35% 20% 0 !important;
		right: 0.35em;
		position: relative;
	}

	.close-button {
		display: inline-block;
		color: #e1e1e1;
		width: var(--close-button-width);
		height: var(--close-button-height);
		fill: none;
	}

	.close-button:hover {
		color: #fafaff;
	}

	.member-name {
		font-weight: bold;
		margin-bottom: 0.5em;
	}

	div#member-bio {
		font-size: 0.9em;
		color: #666;
		margin-bottom: 1em;
		width: 85%;
		position: relative;
		overflow-wrap: break-word;
		word-wrap: break-word;
		display: inline-block;
	}

	.member-joined-date {
		font-size: 0.8em;
		color: #999;
	}

	button#logout-button {
		/* 90% of member card width  */
		width: calc(90% - 2em);
		float: var(--logout-button-align);
		padding-top: var(--logout-button-padding-top);
		margin-left: 10%;
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

	.modal {
		display: none;
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background-color: white;
		padding: 20px;
		border: 1px solid #ccc;
		box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
		z-index: 1000;
	}

	.tooltip {
		position: relative;
		font-size: 0.9em;
		cursor: help;
	}

	.tooltip::before {
		content: '';
		position: absolute;
		top: 110%;
		left: 50%;
		transform: translateX(-50%);
		display: none;
		background-color: #aaa;
		color: #000;
		padding: 0.3em 0.6em;
		border-radius: 4px;
		font-size: 1em;
		white-space: nowrap;
	}

	.tooltip:hover::before {
		display: block;
	}
</style>
