<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { XIcon, MaximizeIcon, EditIcon } from 'svelte-feather-icons';
	import { Button } from '@sveltestrap/sveltestrap';
	import { onMount, onDestroy } from 'svelte';
	import type { Member } from '$lib/types/member.ts';
	import type { NullableString } from '$lib/types/utils';
	import { browser } from '$app/environment';
	import { authStore } from '$stores/members/auth';
	import { followStore, type FollowRequestOut, type FollowResponse } from '$stores/members/follow';
	import UpdateBio from '$components/form/UpdateBio.svelte';
	import { openFilePicker, getMaxFileSize } from '$stores/form/upload';
	import type { CustomHttpError } from '$lib/types/error';
	import ErrorModal from '$components/modal/ErrorModal.svelte';

	const tooltipMessage = 'Change profile picture (max. 400x400px)';
	export let member: Member;
	let currentUser: string = '';
	const jwtToken = localStorage.getItem('jwtToken');

	let isSelfView: boolean = false;
	function splitNullable(input: NullableString, separator: string): string[] {
		if (input.Valid) {
			return input.String.split(separator);
		}
		return [];
	}

	let csrfToken: string | undefined;
	csrfToken = document.cookie
		.split('; ')
		.find((row) => row.startsWith('csrf_'))
		?.split('=')[1];
	let followStatus: FollowResponse;
	let maxFileSize: number;
	let maxFileSizeString: string;
	let errorMessages: CustomHttpError[] = [];

	onMount(async () => {
		isSelfView = (await checkSelfView()) || false;
		maxFileSize = await getMaxFileSize();
	});

	const getFollowStatus = async () => {
		if (!jwtToken) {
			throw new Error('Not logged in');
		}
		followStatus = await followStore.followStatus(jwtToken, member.webfinger);
	};

	onDestroy(() => {
		maxFileSize = 0;
	});

	const followUnfollow = async () => {
		try {
			if (!jwtToken) {
				throw new Error('Not logged in');
			}
			const req: FollowRequestOut = {
				jwtToken: jwtToken,
				target: member.webfinger,
				reblogs: true,
				notify: false,
				CSRFToken: csrfToken || ''
			};

			if (followStatus.status === 'accepted') {
				if (!jwtToken) {
					throw new Error('Not logged in');
				}
				followStatus = await followStore.unfollow(req);
			} else {
				if (!jwtToken) {
					throw new Error('Not logged in');
				}
				followStatus = await followStore.follow(req);
			}
		} catch (error) {
			errorMessages.push({
				message: 'Error following/unfollowing member: ' + error,
				status: 400
			});
			errorMessages = [...errorMessages];
			console.error(error);
			followStatus.status = 'failed';
		}
	};

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

	const checkSelfView = async () => {
		if (jwtToken) {
			const authStatus = await authStore.authenticate(jwtToken);
			currentUser = authStatus.memberName;
			return currentUser === member.memberName;
		}
	};

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
				const response = await fetch('/api/upload/image', {
          method: 'POST', body:
          formData,
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
				const responseJson = await response.json(); 

				const pic_id = responseJson.data.data.pic_id;
				console.log(pic_id);
				const confirmSave = confirm('Save new profile picture?');
if (confirmSave) {
	const response = await fetch(`/api/members/update/${member.memberName}?profile_pic_id=${pic_id}`, {
		method: 'PATCH',
		headers: {
			'Content-Type': 'multipart/form-data',
			Authorization: `Bearer ${jwtToken}`,
			'X-CSRF-Token': csrfToken || ''
		},
		body: JSON.stringify({
			memberName: member.memberName
		})
	});
	if (!response.ok) {
		errorMessages.push({
			message: 'Error updating profile picture',
			status: response.status
		});
		errorMessages = [...errorMessages];
		reject();
	}
} else {
	const response = await fetch(`/api/upload/image/${pic_id}`, {
		method: 'DELETE',
		headers: {
			Authorization: `Bearer ${jwtToken}`,
			'X-CSRF-Token': csrfToken || ''
		}
	});
	if (!response.ok) {
		errorMessages.push({
			message: 'Error deleting profile picture',
			status: response.status
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
		member.bio = event.detail.newBio;
	};

	const cancelFollowReq = async (requestID: number) => {
		try {
			if (!jwtToken || !csrfToken) {
				throw new Error('Not logged in');
			}
			await followStore.cancelFollowRequest(jwtToken, csrfToken, requestID);
		} catch (error) {
			errorMessages.push({
				message: 'Error cancelling follow request: ' + error,
				status: 400
			});
			errorMessages = [...errorMessages];
			console.error(error);
		}
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
			{#if isSelfView}
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
			{/if}
		</div>
	{:else}
		<div class="member-image-container">
			<img
				class="member-image"
				src="/static/avatar-placeholder.png"
				alt="{member.memberName}'s profile picture"
			/>
			{#if isSelfView}
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
			{/if}
		</div>
	{/if}
	<div class="member-name">@{member.memberName}</div>
	<!-- follow button -->
	{#if !isSelfView}
		{#await getFollowStatus()}
			<div class="spinner" />
		{:then}
			{#if followStatus.status === 'failed'}
				<div class="error-message">{$_('error_updating_follow_status')}</div>
			{:else if followStatus.status === 'pending'}
				<Button on:click={() => cancelFollowReq(followStatus.id)}>
					{$_('cancel_follow_request')}
				</Button>
			{/if}
			<Button on:click={followUnfollow}>
				{#if followStatus.status === 'accepted'}
					Unfollow
				{:else if followStatus.status === 'not_found'}
					Follow
				{/if}
			</Button>
		{/await}
	{/if}
	{#if member.bio !== ''}
		<div id="member-bio">{member.bio}</div>
		{#if isSelfView}
			<UpdateBio
				memberName={member.memberName}
				isBioPresent={member.bio !== ''}
				bio={member.bio}
				on:bioUpdated={reloadBio}
			/>
		{:else}
			<UpdateBio
				memberName={member.memberName}
				isBioPresent={member.bio !== ''}
				bio=""
				on:bioUpdated={reloadBio}
			/>
			/>
		{/if}
	{/if}
	<div class="member-joined-date">Joined {member.regdate}</div>

	{#if member.customFields.length > 0}
		{#if member.customFields[0].size > 0}
			<div class="additional-info">
				<table>
					<th>
						{$_('additional-info-header')}{member.memberName}:
					</th>
					{#each Array.from(member.customFields.entries()) as [key, value]}
						<tr>
							<td>{key}</td>
							<td>{value}</td>
						</tr>
					{/each}
				</table>
			</div>
		{/if}
	{/if}
</div>

{#if isSelfView}
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
		--logout-button-align: right;
		--logout-button-padding-top: 0.25em;
		--close-button-align: right;
		--close-button-width: 1.2em;
		--close-button-height: 1.2em;
		--button-bg: #60605190;
		--button-radius: 20%;
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
		margin: 1em 0.3em 1em 0.6em;
		border-radius: var(--member-card-border-radius);
		background-color: var(--member-card-background-color);
		color: var(--member-card-color);
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
		font-weight: 600;
		color: var(--primary-text-color);
		margin-bottom: 0.5em;
	}

	div#member-bio {
		font-size: 0.9em;
		color: var(--tertiary-text-color);
		margin-bottom: 1em;
		width: 85%;
		position: relative;
		overflow-wrap: break-word;
		word-wrap: break-word;
		display: inline-block;
	}

	.member-joined-date {
		font-size: 0.8em;
		color: var(--minor-text-color);
	}

	button#logout-button {
		/* 90% of member card width  */
		width: calc(90% - 2em);
		float: var(--logout-button-align);
		padding-top: var(--logout-button-padding-top);
		margin-left: 10%;
		background: var(--button-bg);
		border-radius: var(--button-radius);
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
