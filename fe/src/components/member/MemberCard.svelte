<script lang="ts">
	import axios from 'axios';
	import type { Member } from '$lib/types/member.ts';
	import type { NullableString } from '$lib/types/utils';
	import { browser } from '$app/environment';
	import { authStore } from '$stores/members/auth';

	const tooltipMessage = 'Change profile picture (max. 400x400px)';
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
			console.error(error);
		}
	};

	let isUploading = false;
	let showModal = false;
	let maxFileSize: number;
	let maxFileSizeString: string;

	const getMaxFileSize = async () => {
		try {
			const response = await axios.get('/api/upload/max-file-size');
			maxFileSize = response.data.maxFileSize;
		} catch (error) {
			maxFileSize = 4 * 1024 * 1024;
		}
	};

	$: {
		maxFileSizeString = `${(maxFileSize / 1024 / 1024).toFixed(2)} MB`;
	}

	const toggleModal = () => {
		showModal = !showModal;
	};
	const jwtToken = localStorage.getItem('jwtToken');

	const openFilePicker = () => {
		if (browser) {
			const fileInput = document.createElement('input');
			fileInput.type = 'file';
			fileInput.accept = 'image/*';
			fileInput.addEventListener('change', handleFileSelection);
			fileInput.click();
		}
	};

	const handleFileSelection = async (e: Event) => {
		return new Promise(async (resolve, reject) => {
			if (browser) {
				let csrfToken: string | undefined;
				csrfToken = document.cookie
					.split('; ')
					.find((row) => row.startsWith('csrf_'))
					?.split('=')[1];

				const fileInput = e.target as HTMLInputElement;
				const file = fileInput.files?.[0];
				if (!file) {
					return;
				}
				await getMaxFileSize();
				if (file.size > maxFileSize) {
					alert('File too large.');
					reject();
				}
				isUploading = true;
				const formData = new FormData();
				formData.append('fileData', file);
				formData.append('imageType', 'profile');
				formData.append('member', member.memberName);
				const response = await axios.post('/api/upload/image', formData, {
					headers: {
						'Content-Type': 'multipart/form-data',
						Authorization: `Bearer ${jwtToken}`,
						Expect: '100-continue',
						'X-CSRF-Token': csrfToken || ''
					}
				});
				if (response.status !== 201) {
					alert('Error uploading file');
					reject();
				}
				console.log('uploaded');
				isUploading = false;
				console.log(response.data);
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
								Expect: '100-continue',
								Authorization: `Bearer ${jwtToken}`,
								'X-CSRF-Token': csrfToken || ''
							}
						}
					);
					if (res.status !== 200) {
						alert('Error saving profile picture');
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
						alert('Error deleting profile picture');
						reject();
					}
				}
				isUploading = false;
				resolve(void 0);
			}
		});
	};

	const toggleSaveButton = (e: Event) => {
		const button = e.target as HTMLButtonElement;
		button.classList.toggle('saved');
		button.classList.toggle('unsaved');
		if (button.classList.contains('saved')) {
			button.innerHTML = '<i class="feather" data-feather="check"></i> Saved';
		} else {
			button.innerHTML = '<i class="feather" data-feather="save"></i> Save';
		}
	};
</script>

<div class="member-card">
	{#if member.profilePic}
		<img class="member-image" src={member.profilePic} alt="{member.memberName}'s profile picture" />
		<button
			aria-label="Change profile picture (max. {maxFileSizeString})"
			id="change-profile-pic-button"
			on:click={openFilePicker}
			><span class="tooltip" aria-label={tooltipMessage} />
			<i class="feather" data-feather="edit-2" />
		</button>
	{:else}
		<img
			class="member-image"
			src="https://www.gravatar.com/avatar/000
    ?d=mp"
			alt="{member.memberName}'s profile picture"
		/>
		<button
			aria-label="Change profile picture (max. {maxFileSizeString})"
			id="change-profile-pic-button"
			on:click={openFilePicker}
			><span class="tooltip" aria-label={tooltipMessage} />
			<i class="feather" data-feather="edit-2" />
		</button>
	{/if}
	<div class="member-name">@{member.memberName}</div>
	{#if member.bio.Valid}
		<div class="member-bio">{member.bio.String}</div>
	{/if}
	<div class="member-joined-date">Joined {regDate}</div>
	Other links and contact info @{member.memberName} has provided:
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
<button aria-label="Logout" on:click={logout} id="logout-button">Logout</button>
{#if showModal}
	<div class="modal">
		<img src={member.profilePic} alt="{member.memberName}'s profile picture" />
		<button on:click={toggleModal} aria-label="Close modal">
			<i class="feather" data-feather="x" />
		</button>
	</div>
{/if}

<style>
	:root {
		--member-card-border-radius: 0.25em;
		--logout-button-align: right;
		--logout-button-padding-top: 3em;
	}

	.member-card {
		border: 1px solid #ccc;
		padding: 1em;
		margin: 1em;
		border-radius: var(--member-card-border-radius);
	}

	.feather {
		width: 0.8em;
		height: 0.8em;
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

	.logout-button {
		float: var(--logout-button-align);
		padding-top: var(--logout-button-padding-top);
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
