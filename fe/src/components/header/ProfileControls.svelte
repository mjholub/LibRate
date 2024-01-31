<script lang="ts">
	import { fly, type TransitionConfig } from 'svelte/transition'
	import { quartOut } from 'svelte/easing';
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

	let mobileMenuAnim: TransitionConfig;

	const toggleMobileMenu = () => {
		showMobileMenu = !showMobileMenu;
	};
</script>

<div class="mobile-menu">
	<button id="menu" title="Menu" on:click={toggleMobileMenu} transition:fly={{
		x: '40%',
		delay: 50,
		easing: quartOut,
		duration: 450
	}}>
		<span class="menu-icon"><MenuIcon /></span>
		<MobileMenu bind:showMobileMenu>
			<span slot="nick" id="menu-nick"><a href="/profiles/{nickname}">@{nickname}</a></span>
			<span slot="settings">
				<button class="text-button" title="Settings" on:click={() => (showSettingsModal = true)}>
					Settings</button
				>
				<Settings bind:showSettingsModal {nickname} />
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
		<Settings bind:showSettingsModal {nickname} />
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
    margin-right: 0.2em;
    margin-block-start: 0.75em;
    right: 1%;
    top: 1%;
		z-index: 150 !important;
	}

	button#settings {
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
