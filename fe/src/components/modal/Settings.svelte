<script lang="ts">
	import { _, locale } from 'svelte-i18n';
	import { Modal, ModalHeader } from '@sveltestrap/sveltestrap';
	import { onMount, onDestroy } from 'svelte';
	import { PenToolIcon, LockIcon } from 'svelte-feather-icons';
	import UxSettings from './tabs/UXSettings.svelte';
	import PrivacySecurity from './tabs/PrivacySecurity.svelte';

	import type { MemberPreferences } from '../../stores/members/prefs.ts';
	import type { TabItem } from '$lib/types/tabs.ts';

	export let showSettingsModal = false;
	export let nickname: string;

	let activeTabIndex = 0;
	const handleClick = (tabIndex: number) => () => (activeTabIndex = tabIndex);

	const handleKeyDown = (event: KeyboardEvent) => {
		if (event.key === 'Shift') {
			event.preventDefault();
		} else if (event.key === 'ArrowRight' || event.key === 'ArrowLeft') {
			const direction = event.key === 'ArrowRight' ? 1 : -1;
			const newIndex = (activeTabIndex + direction + items.length) % items.length;
			activeTabIndex = newIndex;
		}
	};

	onMount(() => {
		window.addEventListener('keydown', handleKeyDown);
	});

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeyDown);
	});

	let items: TabItem[] = [
		{
			label: 'ux',
			index: 0,
			componentName: UxSettings,
			events: {
				themeChanged: handlePrivacySettingsUpdate
			}
		},
		{
			label: 'privsec',
			index: 1,
			componentName: PrivacySecurity,
			events: {
				privacySettingsUpdated: handlePrivacySettingsUpdate
			}
		}
	];

	function handlePrivacySettingsUpdate(event: CustomEvent) {
		prefs.privsec = event.detail.newSettings;
	}

	// initialize with default values
	export let prefs: MemberPreferences = {
		ux: {
			locale: $locale,
			theme: 'default',
			rating_scale_lower: 0,
			rating_scale_upper: 10
		},
		privsec: {
			auto_accept_follow: true,
			locally_searchable: true,
			robots_searchable: false,
			blur_nsfw: true,
			searchable_to_federated: true,
			message_autohide_words: [],
			muted_instances: []
		}
	};

	const toggle = () => (showSettingsModal = !showSettingsModal);
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://cdn.jsdelivr.net/npm/bootstrap@4.3.2/dist/css/bootstrap.min.css"
	/>
</svelte:head>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
<div>
	<Modal isOpen={showSettingsModal} {toggle} size="xl">
		<ModalHeader {toggle} class="settings-text" id="settings-text">{$_('settings')}</ModalHeader>
		<ul>
			{#each items as item}
				<li class={activeTabIndex === item.index ? 'active' : ''}>
					<!-- svelte-ignore a11y-click-events-have-key-events -->
					<!-- handled by the onMount event listener -->
					<span on:click={handleClick(item.index)} tabindex="-1" role="tabpanel">
						{#if item.label === 'ux'}
							<PenToolIcon />
							{$_('ux')}
						{:else if item.label === 'privsec'}
							<LockIcon />
							{$_('privsec')}
						{/if}
					</span>
				</li>{/each}
		</ul>
		<p class="info-msg">{$_('key-combination-infomsg-tabs')}</p>
		<div class="box">
			{#if activeTabIndex === 0}
				<UxSettings />
			{:else}
				<PrivacySecurity
					on:privacySettingsUpdated={handlePrivacySettingsUpdate}
					memberName={nickname}
				/>
			{/if}
		</div>

		<button on:click={() => (showSettingsModal = false)}>{$_('close')}</button>
	</Modal>
</div>

<style>
	.box {
		margin-bottom: 10px;
		padding: 40px;
		border: 1px solid #dee2e6;
		border-radius: 0 0 0.5rem 0.5rem;
		border-top: 0;
	}
	ul {
		display: flex;
		flex-wrap: wrap;
		padding-left: 0;
		margin-bottom: 0;
		list-style: none;
		border-bottom: 1px solid #dee2e6;
	}
	li {
		margin-bottom: -1px;
	}

	span {
		border: 1px solid transparent;
		border-top-left-radius: 0.25rem;
		border-top-right-radius: 0.25rem;
		display: block;
		padding: 0.5rem 1rem;
		cursor: pointer;
	}

	span:hover {
		border-color: #e9ecef #e9ecef #dee2e6;
	}

	li.active > span {
		color: #495057;
		background-color: #fff;
		border-color: #dee2e6 #dee2e6 #fff;
	}

	button {
		display: block;
	}
</style>
