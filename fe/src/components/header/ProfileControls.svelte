<script lang="ts">
	import Settings from '$components/modal/Settings.svelte';
	import MobileMenu from '$components/modal/MobileMenu.svelte';
	import { PowerIcon, MenuIcon, SettingsIcon } from 'svelte-feather-icons';
	import { browser } from '$app/environment';
	import { authStore } from '$stores/members/auth';
	let showSettingsModal = false;
	let showMobileMenu = false;
	export let nickname: string;

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

	const toggleMobileMenu = () => {
		showMobileMenu = !showMobileMenu;
	};
</script>

<div class="mobile-menu">
	<button id="menu" title="Menu" on:click={toggleMobileMenu}>
		<span class="menu-icon"><MenuIcon /></span>
		<MobileMenu bind:showMobileMenu>
			<span slot="nick" id="menu-nick"><a href="/profiles/{nickname}">@{nickname}</a></span>
			<span slot="settings">
				<button class="text-button" title="Settings" on:click={() => (showSettingsModal = true)}>
					Settings</button
				>
				<Settings bind:showSettingsModal>
					<h2 slot="settings" id="settings-text">Settings</h2>
				</Settings>
			</span>
			<span slot="logout">
				<button class="text-button" on:click={logout}>Log out </button>
			</span></MobileMenu
		>
	</button>
</div>
<div class="profile-controls">
	<span class="profile-nickname">
		<a href="/profiles/{nickname}">@{nickname}</a>
	</span>
	<button id="settings" title="Settings" on:click={() => (showSettingsModal = true)}>
		<span class="settings-icon"><SettingsIcon /></span>
		<Settings bind:showSettingsModal>
			<h2 slot="settings" id="settings-text">Settings</h2>
		</Settings>
		<button id="logout" title="Logout" on:click={logout}>
			<PowerIcon />
		</button>
	</button>
</div>

<style>
  :root {
    --link-color: #c68
  --link-hover-color: #d8a
  }

  @media (max-width: 768px) {
    .profile-controls {
      display: none !important;
    }
  }

  @media (min-width: 768px) {
    .mobile-menu {
      display: none;
    }
  }

  .mobile-menu:active {
    display: flex;
    position: fixed;
    height: 45%;
    justify-content: center;
  }

  .profile-controls {
  display: flex;
  align-items: center;
  justify-content: center;
  }

  .settings-icon {
    padding-right: 0.5em;
    display: inline-flex;
  } 

  button#settings, button#logout
  {
    display: inline-flex;
    align-items: center;
    margin-block-start: 0.5em;
  }

	button#menu {
		display: block;
    position: fixed;
    margin-right: 0;
    margin-block-start: 0.75em;
    right: 1%;
    top: 1%;
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

	.text-button {
		display: block;
		position: relative;
		width: 100%;
		margin-block-end: 0.6em;
		font-weight: 600;
		font-size: 4mm;
	}

</style>
