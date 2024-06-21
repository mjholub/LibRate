<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { SaveIcon, CheckIcon, Edit2Icon, XIcon } from 'svelte-feather-icons';

	const dispatch = createEventDispatcher();

	export let memberName: string;
	export let isBioPresent: boolean;
	export let bio: string;
	$: inputActive = false;
	$: bioSaved = false;
	const jwtToken = localStorage.getItem('jwtToken') || '';

	let errorMessages: string[] = [];

	const getBioComponents = () => {
		const bioInput = document.getElementById('bio-input') as HTMLTextAreaElement;
		const editBio = document.getElementById('edit-bio') as HTMLAnchorElement;
		const saveBio = document.getElementById('save-bio-button') as HTMLButtonElement;
		const cancelBio = document.getElementById('cancel-bio-button') as HTMLButtonElement;
		const bioContents = document.getElementById('member-bio') as HTMLDivElement;
		return { bioInput, editBio, saveBio, cancelBio, bioContents };
	};

	const showBioInput = () => {
		const { bioInput, editBio, saveBio, cancelBio, bioContents } = getBioComponents();
		if (bioInput && editBio) {
			bioInput.style.display = 'block';
			editBio.style.display = 'none';
			saveBio.style.display = 'contents';
			cancelBio.style.display = 'contents';
			if (bioContents) {
				bioContents.style.display = 'none';
			}
			bioInput.focus();
			inputActive = true;
		}
	};

	const hideBioInput = () => {
		const { bioInput, editBio, saveBio, cancelBio, bioContents } = getBioComponents();
		if (bioInput && editBio) {
			bioInput.style.display = 'none';
			editBio.style.display = 'contents';
			saveBio.style.display = 'none';
			cancelBio.style.display = 'none';
			if (bioContents) {
				bioContents.style.display = 'inline-block';
			}
			inputActive = false;
		}
	};

	const saveBio = async () => {
		const { bioInput, editBio, saveBio, cancelBio } = getBioComponents();
		if (bioInput && editBio && cancelBio && saveBio) {
			const csrfToken = document.cookie
				.split('; ')
				.find((row) => row.startsWith('csrf_'))
				?.split('=')[1];
			const formData = new FormData();
			formData.append('bio', bio);
			const response = await fetch(`/api/members/update/${memberName}`, {
        method: 'PATCH',
				headers: {
					'Content-Type': 'multipart/form-data',
					Authorization: `Bearer ${jwtToken}`,
					'X-CSRF-Token': csrfToken || ''
				},
        body: formData
			});
			if (response.status !== 200) {
        const errorMessage = await response.text();
				errorMessages.push(`Error saving bio: ${errorMessage} (${response.status})`);
				errorMessages = [...errorMessages]
			}

			bioSaved = true;
			bio = bioInput.value;
			dispatch('bioUpdated', {
				newBio: bio
			});
			hideBioInput();
		}
	};

	//	$: bio = bio || '';
</script>

{#if isBioPresent}
	<button id="edit-bio" on:click={showBioInput} on:keypress={showBioInput} aria-label="Edit bio">
		<span class="icon" id="edit-icon">
			<Edit2Icon />
		</span>
	</button>
	<button
		aria-label="Close"
		on:click={hideBioInput}
		on:keypress={hideBioInput}
		id="cancel-bio-button"
	>
		<span class="close-icon">
			<XIcon />
		</span>
	</button>
	<textarea
		id="bio-input"
		bind:value={bio}
		on:input={() => {
			bioSaved = false;
		}}
	/>
{:else}
	<a id="edit-bio" on:click={showBioInput} on:keypress={showBioInput} role="button" tabindex="0"
		>Add bio</a
	>
	<button
		aria-label="Close"
		on:click={hideBioInput}
		on:keypress={hideBioInput}
		id="cancel-bio-button"
	>
		<span class="close-icon">
			<XIcon />
		</span>
	</button>

	<textarea
		id="bio-input"
		placeholder="Enter bio here"
		bind:value={bio}
		on:input={() => {
			bioSaved = false;
		}}
	/>
{/if}
<button aria-label="Save bio" on:click={saveBio} on:keypress={saveBio} id="save-bio-button">
	{#if bioSaved && inputActive}
		<span class="save-icon" aria-label="Changes saved!">
			<CheckIcon />
		</span>
	{:else}
		<span class="save-icon" aria-label="Save changes">
			<SaveIcon />
		</span>
	{/if}
</button>

<style>
	:root {
		--edit-icon-color: #ffcbcc;
		--edit-icon-color-hover: calc(var(--icon-color) + 0.2);
		--button-bg: #606051;
		--button-radius: 4px;
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
		display: contents;
		position: relative;
		bottom: 1%;
		left: 1%;
		padding: 0.1em 0.2em;
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
	#cancel-bio-button {
		margin-left: 93%;
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

	span.close-icon {
		margin-left: 92%;
	}

	span.save-icon {
		position: relative;
		margin-bottom: 3%;
		margin-left: 1%;
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
