<script lang="ts">
	import Settings from '$components/modal/Settings.svelte';
	import { PowerIcon, SettingsIcon } from 'svelte-feather-icons';
	import { browser } from '$app/environment';
	import { authStore } from '$stores/members/auth';
	const tooltipText = 'Logout';
	export let nickname: string;
	let showSettingsModal = false;

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
</script>

<div class="profile-controls">
	<span class="profile-nickname">
		<a href="/profiles/{nickname}">@{nickname}</a>
	</span>
	<button id="settings" title="Settings" on:click={() => (showSettingsModal = true)}>
		<SettingsIcon />
		<Settings bind:showSettingsModal>
			<h2 slot="settings" id="settings-text">Settings</h2>
		</Settings>
		<button id="logout" title={tooltipText} on:click={logout}>
			<PowerIcon />
		</button>
	</button>
</div>

<style>
  :root {
    --link-color: #c68
    --link-hover-color: #d8a
  }
  .profile-controls {
  display: flex;
  align-items: center;
  }

  button#settings, button#logout
  {
    display: inline-flex;
    align-items: center;
    margin-block-start: 0.5em;
  }

	button#settings h2#settings-text {
		margin-left: 1em;
		padding: 0;
		display: inline-flex;
	}

	.profile-nickname {
  display: inline-block;
  max-width: 60%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 0.6em;
	}

  .profile-nickname > a {
    color: var(--link-color);
    text-decoration: none;
    font-weight: 600;
  }

  button#logout {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--link-color);
    padding: 0;
    margin: 0;
    font-size: 1em;
  }

  button#logout:hover {
    color: var(--link-hover-color);
  }

	.profile-nickname > a:hover {
  text-decoration: underline;
  color: var(--link-hover-color);
  font-weight: 800;
	}
</style>
