<script lang="ts">
	import axios from 'axios';
	import { XCircleIcon, SaveIcon, CheckIcon, Edit2Icon, XIcon } from 'svelte-feather-icons';

	export let memberName: string;
	export let isBioPresent: boolean;
	export let bio: string;
	//export let inputActive: boolean;
	let inputActive = false;
	let bioSaved = false;
	const jwtToken = localStorage.getItem('jwtToken') || '';

	const errorMessages: string[] = [];

	// TODO: use this to reactively hide the bio contents when editing is active

	const showBioInput = () => {
		const bioInput = document.getElementById('bio-input') as HTMLTextAreaElement;
		const editBio = document.getElementById('edit-bio') as HTMLAnchorElement;
		const saveBio = document.getElementById('save-bio-button') as HTMLButtonElement;
		const cancelBio = document.getElementById('cancel-bio-button') as HTMLButtonElement;
		if (bioInput && editBio) {
			bioInput.style.display = 'block';
			editBio.style.display = 'none';
			saveBio.style.display = 'inline-block';
			cancelBio.style.display = 'inline-block';
			bioInput.focus();
			inputActive = true;
		}
	};

	const saveBio = async () => {
		const bioInput = document.getElementById('bio-input') as HTMLTextAreaElement;
		const editBio = document.getElementById('edit-bio') as HTMLAnchorElement;
		const cancelBio = document.getElementById('cancel-bio-button') as HTMLButtonElement;
		const saveBio = document.getElementById('save-bio-button') as HTMLButtonElement;
		if (bioInput && editBio && cancelBio && saveBio) {
			const csrfToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('csrf_'))
				?.split('=')[1];
			const bio = bioInput.value;
			const formData = new FormData();
			formData.append('bio', bio);
			const response = await axios.patch(`/api/members/update/${memberName}`, formData, {
				headers: {
					'Content-Type': 'multipart/form-data',
					Authorization: `Bearer ${jwtToken}`,
					Expect: '100-continue',
					'X-CSRF-Token': csrfToken || ''
				}
			});
			if (response.status !== 200) {
				errorMessages.push('Error saving bio');
			}

			bioSaved = true;
			inputActive = false;
		}
	};
</script>

{#if isBioPresent}
	<button id="edit-bio" on:click={showBioInput} on:keypress={showBioInput} aria-label="Edit bio">
		<span class="icon" id="edit-icon">
			<Edit2Icon />
		</span>
	</button>
	<textarea id="bio-input">
		{bio}
	</textarea>
{:else}
	<a id="edit-bio" on:click={showBioInput} on:keypress={showBioInput} role="button" tabindex="0"
		>Add bio</a
	>
	<textarea id="bio-input" placeholder="Enter bio here" />
{/if}
<button
	aria-label="Close"
	on:click={showBioInput}
	on:keypress={showBioInput}
	id="cancel-bio-button"
>
	<span class="close-icon">
		<XIcon />
	</span>
</button>
<button aria-label="Save bio" on:click={saveBio} on:keypress={saveBio} id="save-bio-button">
	{#if bioSaved}
		<span class="icon">
			<CheckIcon />
		</span>
	{:else}
		<span class="icon">
			<SaveIcon />
		</span>
	{/if}
</button>
<button
	aria-label="Cancel changes"
	on:click={showBioInput}
	on:keypress={showBioInput}
	id="cancel-bio-button"
>
	<span class="icon">
		<XCircleIcon />
	</span>
</button>

<style>
	:root {
		--edit-icon-color: #ffcbcc;
		--edit-icon-color-hover: calc(var(--icon-color) + 0.2);
		--button-bg: #606051;
		--button-radius: 20%;
		--bio-input-font-size: 11pt;
		--bio-input-height: 6em;
	}
	#bio-input {
		display: none;
		width: 100%;
		height: var(--bio-input-height);
		font-size: var(--bio-input-font-size);
		resize: none;
	}
	#edit-bio {
		color: var(--edit-icon-color) !important;
		display: flex;
		position: relative;
		bottom: 15%;
		left: 90%;
		width: 1.75em !important;
		height: 1.75em !important;
		background-color: var(--button-bg);
		border-radius: var(--button-radius);
	}
	#save-bio-button,
	#cancel-bio-button {
		display: none;
		height: 1.75em !important;
		width: 1.75em !important;
		margin-right: 0.5em;
	}
	#save-bio-button:hover,
	#cancel-bio-button:hover {
		cursor: pointer;
	}
	#save-bio-button:focus,
	#cancel-bio-button:focus {
		outline: none;
	}
	#save-bio-button:active,
	#cancel-bio-button:active {
		transform: scale(0.9);
	}
	#edit-bio {
		color: var(--text-color);
	}
	#edit-bio:hover {
		color: var(--text-color-hover);
		cursor: pointer;
	}
	#edit-bio:focus {
		outline: none;
	}
	#edit-bio:active {
		transform: scale(0.9);
	}

	.icon {
		color: var(--edit-icon-color);
		display: block;
		fill: none;
		width: 1em;
		height: 1em;
		background-color: var(--button-bg);
		border-radius: var(--button-radius);
	}

	.icon:hover {
		color: var(--edit-icon-color-hover);
		cursor: pointer;
	}

	.icon:focus {
		outline: none;
	}

	.icon:active {
		transform: scale(0.9);
	}

	.close-icon {
		color: #6f0000;
		width: 0.8em;
		height: 0.8em;
	}

	.icon#edit-icon {
		display: inherit !important;
		width: 1em;
		height: 1em;
		fill: none;
		color: var(--edit-icon-color);
		margin-top: -0.35em !important;
		padding: 0 35% 20% 0 !important;
		position: relative;
	}
</style>
