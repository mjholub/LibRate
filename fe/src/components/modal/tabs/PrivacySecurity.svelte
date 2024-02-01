<script defer lang="ts">
	// @ts-ignore
	import Tags from 'svelte-tags-input';
	import { PasswordMeter } from 'password-meter';
	import { _ } from 'svelte-i18n';
	import type { PrivacySecurityPreferences } from '$stores/members/prefs';
	import { createEventDispatcher } from 'svelte';
	import { authStore, type PasswordUpdateInput } from '$stores/members/auth';
	import axios from 'axios';
	import PasswordInput from '$components/form/PasswordInput.svelte';
	let errorMessages: string[] = [];
	let confirmMutingInstance = false;

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

	const comparePasswords = async (password: string, passwordConfirm: string) => {
		if (password !== passwordConfirm) {
			errorMessages.push('Passwords do not match');
			errorMessages = [...errorMessages];
		} else {
			passwordUpdateData.new = password;
			passwordUpdateData.newConfirm = passwordConfirm;
		}
	};

	const exportData = async;

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
	<label class="settings-label" for="auto-accept-follow">{$_('auto_accept_follow')} </label>
	<input type="checkbox" id="auto-accept-follow" bind:value={settings.auto_accept_follow} />
	<div class="settings-text-input">
		<label for="muted-instances">{$_('muted')} {$_('instances')}</label>
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

		<label class="settings-label" for="federated searchable">
			{$_('known_network')}
		</label>
		<input
			type="checkbox"
			id="federated-searchable"
			bind:value={settings.searchable_to_federated}
		/>

		<label class="settings-label" for="locally-searchable">
			{$_('local_accounts')}
		</label>
		<input type="checkbox" id="locally-searchable" bind:value={settings.locally_searchable} />

		<label class="settings-label" for="robots-searchable">
			{$_('searchable_to_robots')}
		</label>
		<input type="checkbox" id="robots-searchable" bind:value={settings.robots_searchable} />

		<hr />
		<div class="data-export">
			<h3 class="settings-section-descriptor">{$_('data_export')}</h3>
			<select id="select-export-format" bind:value={exportFormat}
				>{$_('export_format')}
				{#each ['json', 'csv', 'sql'] as format}
					<option value={format}>{format.toUpperCase}</option>
				{/each}
			</select>
			<button type="submit" class="submit-button" on:click={exportData}>{$_('export')}</button>
		</div>

		<div class="danger-zone">
			<h3 class="settings-section-descriptor">{$_('danger_zone')}</h3>
			<div class="passwd-change">
				<details class="danger-zone-details">
					<summary>{$_('change_password')}</summary>
					<label for="current-password">{$_('current_password')}</label>
					<input type="password" id="current-password" bind:value={passwordUpdateData.current} />
					<label for="new-password">{$_('new_password')}</label>
					<PasswordInput
						bind:value={passwordUpdateData.new}
						id="new-password"
						onInput={async () => {
							checkEntropy(passwordUpdateData.new);
						}}
						{showPassword}
						{toggleObfuscation}
					/>
					<label for="confirm-new-password">{$_('confirm_new_password')}</label>
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
					<label for="password">{$_('password')}</label>
					<PasswordInput
						onInput={async () => void 0}
						bind:value={deletionPassword}
						id="password"
						{showPassword}
						{toggleObfuscation}
					/>
					<label for="confirm-password">{$_('confirm_password')}</label>
					<PasswordInput
						onInput={async () => comparePasswords(deletionPassword, deletionPasswordConfirm)}
						bind:value={deletionPasswordConfirm}
						id="confirm-password"
						{showPassword}
						{toggleObfuscation}
					/>
				</details>
			</div>
		</div>
	</div>

	<hr />
</form>
