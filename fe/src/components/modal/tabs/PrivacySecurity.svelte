<script defer lang="ts">
	// @ts-ignore
	import Tags from 'svelte-tags-input';
	import { Button, Modal, ModalBody, ModalFooter, ModalHeader } from '@sveltestrap/sveltestrap';

	import { PasswordMeter } from 'password-meter';
	import { _ } from 'svelte-i18n';
	import type { PrivacySecurityPreferences } from '$stores/members/prefs';
	import { createEventDispatcher } from 'svelte';
	import {
		memberStore,
		type DataExportFormat,
		type DataExportRequest
	} from '$stores/members/getInfo';
	import { authStore, type PasswordUpdateInput } from '$stores/members/auth';
	import axios from 'axios';
	import PasswordInput from '$components/form/PasswordInput.svelte';
	let errorMessages: string[] = [];
	let confirmMutingInstance = false;
	let confirmDialogOpen = false;

	// TODO: actual logic to fetch and cache the known network as suggestions setting for muted instances
	const knownInstances = ['bookwyrm.social'];

	$: settingsSaved = false;

	export let memberName: string;
	let showPassword = false;
	let strength: number;
	let passwordStrength = '';
	let timeoutId: number | undefined;

	let deletionPassword = '';
	let deletionPasswordConfirm = '';
	// SQL: DDL
	let exportFormat: DataExportFormat = 'json';

	// TODO: fetch the current settings
	let settings: PrivacySecurityPreferences = {
		auto_accept_follow: true,
		locally_searchable: true,
		robots_searchable: false,
		blur_nsfw: true,
		searchable_to_federated: true,
		message_autohide_words: [],
		muted_instances: []
	};

	let passwordUpdateData = {
		current: '',
		new: '',
		newConfirm: ''
	};

	const dispatch = createEventDispatcher();

	const toggleConfirmDialog = () => {
		confirmDialogOpen = !confirmDialogOpen;
	};

	const toggleObfuscation = () => {
		showPassword = !showPassword;
	};

	const csrfToken = document.cookie
		.split('; ')
		.find((row) => row.startsWith('csrf_'))
		?.split('=')[1];
	const jwtToken = localStorage.getItem('jwtToken') || '';

	const checkEntropy = async (password: string) => {
		if (timeoutId) {
			window.clearTimeout(timeoutId);
		}

		timeoutId = window.setTimeout(async () => {
			try {
				strength = new PasswordMeter().getResult(password).score;
				passwordStrength = strength > 135 ? 'Password is strong enough' : `${strength / 2.9} bits`;
			} catch (error) {
				errorMessages.push('Password is not strong enough or error occurred');
				errorMessages = [...errorMessages];
			}
		}, 300);
	};

	const deleteAccount = async () => {
		const input: PasswordUpdateInput = {
			csrfToken: csrfToken || '',
			jwtToken: jwtToken || '',
			old: deletionPassword,
			new: deletionPasswordConfirm
		};
		try {
			await authStore.deleteAccount(input);
		} catch (error: any) {
			errorMessages.push(`Error deleting account: ${error.message} (${error.status})`);
			errorMessages = [...errorMessages];
		}
	};

	const comparePasswords = async (password: string, passwordConfirm: string) => {
		if (password !== passwordConfirm) {
			errorMessages.push('Passwords do not match');
			errorMessages = [...errorMessages];
		} else {
			passwordUpdateData.new = password;
			passwordUpdateData.newConfirm = passwordConfirm;
		}
	};

	const exportData = async () => {
		const input: DataExportRequest = {
			jwtToken: jwtToken,
			target: exportFormat
		};
		try {
			await memberStore.exportData(input);
		} catch (error: any) {
			errorMessages.push('Error while exporting data');
			errorMessages = [...errorMessages];
		}
	};

	const updatePassword = async () => {
		const input: PasswordUpdateInput = {
			csrfToken: csrfToken || '',
			jwtToken: jwtToken || '',
			old: passwordUpdateData.current,
			new: passwordUpdateData.new
		};
		try {
			await authStore.changePassword(input);
		} catch (error: any) {
			errorMessages.push(`Error updating password: ${error.message} (${error.status})`);
			errorMessages = [...errorMessages];
		}
	};

	const settingsUpdate = async () => {
		{
			const res = await axios.patch(`/api/members/update/${memberName}/preferences`, settings, {
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${jwtToken}`,
					'X-CSRF-Token': csrfToken || ''
				}
			});
			if (res.status !== 200) {
				errorMessages.push(`Error updating settings: ${res.data.message} (${res.status})`);
				errorMessages = [...errorMessages];
			}
			settingsSaved = true;

			dispatch('privacySettingsUpdated', {
				newSettings: settings
			});
		}
	};
</script>

<form id="privacy-settings" on:submit={settingsUpdate}>
	<h3 class="settings-section-descriptor">{$_('interactions')}</h3>
	<span class="settings-option">
		<label class="settings-label" for="auto-accept-follow">{$_('auto_accept_follow')} </label>
		<input type="checkbox" id="auto-accept-follow" bind:value={settings.auto_accept_follow} />
	</span>
	<div class="settings-text-input">
		<h5 id="muted-instances">{$_('muted')} {$_('instances')}</h5>
		<label for="confirm-muting-instance">{$_('require_instance_mute_confirmation')}</label>
		<input type="checkbox" bind:value={confirmMutingInstance} />
		{#if confirmMutingInstance}
			<Tags
				bind:tags={settings.muted_instances}
				onlyUnique={true}
				autoComplete={knownInstances}
				onlyAutoComplete={true}
				onTagAdded={confirm('Really mute this instance?')}
			/>
		{:else}
			<!-- TODO: cosider using element.setAttribute -->
			<Tags
				bind:tags={settings.muted_instances}
				onlyUnique={true}
				autoComplete={knownInstances}
				onlyAutoComplete={true}
			/>
		{/if}
		<label for="autohide-words">
			{$_('message_autohide_words')}
		</label>
		<Tags bind:tags={settings.message_autohide_words} onlyUnique={true} />

		<h3 class="settings-section-descriptor">{$_('who_can_search_my_profile')}</h3>

		<span class="settings-option">
			<label class="settings-label" for="federated searchable">
				{$_('known_network')}
			</label>
			<input
				type="checkbox"
				id="federated-searchable"
				bind:value={settings.searchable_to_federated}
			/>
		</span>

		<span class="settings-option">
			<label class="settings-label" for="locally-searchable">
				{$_('local_accounts')}
			</label>
			<input type="checkbox" id="locally-searchable" bind:value={settings.locally_searchable} />
		</span>

		<span class="settings-option">
			<label class="settings-label" for="robots-searchable">
				{$_('searchable_to_robots')}
			</label>
			<input type="checkbox" id="robots-searchable" bind:value={settings.robots_searchable} />
		</span>

		<hr />
		<div class="data-export">
			<h3 class="settings-section-descriptor">{$_('data_export')}</h3>
			<label for="select-export-format">{$_('select_format')} </label>
			<select id="select-export-format" bind:value={exportFormat}>
				{#each ['json', 'csv'] as format, i}
					<option value={format}>{['JSON', 'CSV'][i]}</option>
				{/each}
			</select>
			<button type="submit" class="submit-button" on:click={exportData}>{$_('export')}</button>
		</div>

		<div class="danger-zone">
			<h3 class="settings-section-descriptor">{$_('danger_zone')}</h3>
			<div class="passwd-change">
				<details class="danger-zone-details">
					<summary>{$_('change_password')}</summary>
					<label for="current-password" class="text-input-label">{$_('current_password')}</label>
					<input type="password" id="current-password" bind:value={passwordUpdateData.current} />
					<label for="new-password" class="text-input-label">{$_('new_password')}</label>
					<PasswordInput
						bind:value={passwordUpdateData.new}
						id="new-password"
						onInput={async () => {
							checkEntropy(passwordUpdateData.new);
						}}
						{showPassword}
						{toggleObfuscation}
					/>
					<label for="confirm-new-password" class="text-input-label"
						>{$_('confirm_new_password')}</label
					>
					<PasswordInput
						onInput={async () => {
							comparePasswords(passwordUpdateData.new, passwordUpdateData.newConfirm);
						}}
						bind:value={passwordUpdateData.newConfirm}
						id="confirm-new-password"
						{showPassword}
						{toggleObfuscation}
					/>
					<div class="submit-cancel-container">
						<button type="submit" class="submit-button" on:click={updatePassword}
							>{$_('submit')}</button
						>
						<button
							type="button"
							class="cancel-button"
							on:click={() => {
								passwordUpdateData = {
									current: '',
									new: '',
									newConfirm: ''
								};
							}}>{$_('cancel')}</button
						>
					</div>
				</details>
				<details class="danger-zone-details">
					<summary>{$_('delete_account')}</summary>
					<label for="password" class="text-input-label">{$_('password')}</label>
					<PasswordInput
						onInput={async () => void 0}
						bind:value={deletionPassword}
						id="password"
						{showPassword}
						{toggleObfuscation}
					/>
					<label for="confirm-password" class="text-input-label">{$_('confirm_password')}</label>
					<PasswordInput
						onInput={async () => comparePasswords(deletionPassword, deletionPasswordConfirm)}
						bind:value={deletionPasswordConfirm}
						id="confirm-password"
						{showPassword}
						{toggleObfuscation}
					/>
					<div class="submit-cancel-container">
						<button type="submit" class="submit-button" on:click={toggleConfirmDialog}
							>{$_('confirm')}</button
						>
						<button
							type="button"
							class="cancel-button"
							on:click={() => {
								deletionPassword = '';
								deletionPasswordConfirm = '';
							}}>{$_('cancel')}</button
						>
						<Modal isOpen={confirmDialogOpen} toggle={toggleConfirmDialog}>
							<ModalHeader>{$_('confirm')}</ModalHeader>
							<ModalBody>
								{$_('delete_account_confirm')}
							</ModalBody>
							<ModalFooter>
								<Button color="danger" onClick={deleteAccount}>{$_('confirm')}</Button>
								<Button color="secondary" onClick={toggleConfirmDialog}>{$_('cancel')}</Button>
							</ModalFooter>
						</Modal>
					</div>
				</details>
			</div>
		</div>
	</div>

	<hr />
</form>

<style lang="scss">
	input[type='checkbox'] {
		float: left;
	}

	.settings-option {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: space-between;
	}

	#muted-instances {
		margin-top: 3%;
	}

	#select-export-format {
		display: block;
		margin-bottom: 1%;
	}

	.submit-cancel-container {
		display: block;
		padding: 2% 0;
	}

	label {
		margin-left: 1%;
		padding: 1% 0;
		display: block;
	}

	.text-input-label {
		padding: 1% 0;
		display: block;
	}
</style>
