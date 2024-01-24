<script lang="ts">
	import Search from '$components/utility/Search.svelte';
	import ProfileControls from './ProfileControls.svelte';
	import LangSwitch from '$components/utility/LangSwitch.svelte';
	import { _ } from 'svelte-i18n';
	import {
		DropdownToggle,
		Dropdown,
		DropdownMenu,
		DropdownItem,
		Nav,
		NavbarBrand
	} from '@sveltestrap/sveltestrap';

	let musicOpen = false,
		filmOpen = false,
		isOpen = false,
		booksOpen = false;

	export let authenticated: boolean = false;
	export let nickname: string = '';
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css"
	/>
</svelte:head>

<div class="header">
	<div class="dropdown-left">
		<span class="hostname">
			<NavbarBrand href="/">{window.location.host}&nbsp;&nbsp;</NavbarBrand>
		</span>
		<Nav class="ms-auto" navbar tabs={true} card={true}>
			<Dropdown group={true} direction={'down'}>
				<DropdownToggle caret>{$_('discover')}</DropdownToggle>
				<DropdownMenu id="grouped-discover" right end>
					<Dropdown isOpen={musicOpen} toggle={() => (musicOpen = !musicOpen)} direction={'right'}>
						<DropdownToggle nav caret>{$_('music')}</DropdownToggle>
						<DropdownMenu id="music-discover">
							<DropdownItem><a href="/genres/music">{$_('genres')}</a></DropdownItem>
							<DropdownItem><a href="/releases/music">{$_('releases')}</a></DropdownItem>
						</DropdownMenu>
					</Dropdown>
					<DropdownItem divider />
					<Dropdown isOpen={filmOpen} toggle={() => (filmOpen = !filmOpen)} direction={'right'}>
						<DropdownToggle caret nav>Film</DropdownToggle>
						<DropdownMenu id="film-discover">
							<DropdownItem><a href="/genres/film">{$_('genres')}</a></DropdownItem>
							<DropdownItem><a href="/releases/film">{$_('releases')}</a></DropdownItem>
						</DropdownMenu>
					</Dropdown>
					<DropdownItem divider />
					<Dropdown isOpen={booksOpen} toggle={() => (booksOpen = !booksOpen)} direction={'right'}>
						<DropdownToggle caret nav>{$_('books')}</DropdownToggle>
						<DropdownMenu id="book-discover">
							<DropdownItem><a href="/genres/books">{$_('genres')}</a></DropdownItem>
							<DropdownItem><a href="/releases/books">{$_('releases')}</a></DropdownItem>
							<DropdownItem><a href="/authors">{$_('authors')}</a></DropdownItem>
						</DropdownMenu>
					</Dropdown>
				</DropdownMenu>
			</Dropdown>
		</Nav>
	</div>

	{#if authenticated}
		<!-- leave space for profile controls -->
		<span class="search-with-space">
			<Search />
		</span>
		<span class="add-media">
			<a href="/form/media/add">{$_('add_media')}</a>
		</span>
		<span class="profile-controls">
			<ProfileControls {nickname} />
		</span>
	{:else}
		<span class="search">
			<Search />
		</span>
		<!-- when logged in, language switching available via settings modal -->
		<span class="lang-switch">
			<LangSwitch />
		</span>
	{/if}
</div>

<style lang="scss">
	$justify-header-content: left;

	.hostname {
		font-weight: 300;
		font-size: 0.6em;
		word-break: break-all;
		position: relative;
		width: 5% !important;
		display: flex;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.2em 0.6em 0 0;
		position: relative;
		z-index: 0;
	}

	.header > * {
		height: 2.5em;
		padding: 0.5em;
		flex-direction: row;
		display: grid;
		transition: transform 500ms ease-in-out;
		justify-content: $justify-header-content;
	}

	span.search-with-space {
		flex-grow: 3;
		min-width: 30%;
	}

	span.profile-controls {
		display: inline-flex;
		flex-shrink: 0;
		padding-left: 3.2em;
		align-items: center;
		justify-content: center;
		padding-block-end: 0.5em;
	}

	.dropdown-left {
		display: flex;
		position: relative;
		align-self: start;
		justify-content: space-between;
		font-weight: 600;
	}
</style>
