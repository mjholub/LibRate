<script lang="ts">
	import Search from '$components/utility/Search.svelte';
	import ProfileControls from './ProfileControls.svelte';
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
				<DropdownToggle caret>Discover</DropdownToggle>
				<DropdownMenu id="grouped-discover" right end>
					<Dropdown isOpen={musicOpen} toggle={() => (musicOpen = !musicOpen)} direction={'right'}>
						<DropdownToggle nav caret>Music</DropdownToggle>
						<DropdownMenu id="music-discover">
							<DropdownItem><a href="/genres/music">Genres</a></DropdownItem>
							<DropdownItem><a href="/releases/music">Releases</a></DropdownItem>
						</DropdownMenu>
					</Dropdown>
					<DropdownItem divider />
					<Dropdown isOpen={filmOpen} toggle={() => (filmOpen = !filmOpen)} direction={'right'}>
						<DropdownToggle caret nav>Film</DropdownToggle>
						<DropdownMenu id="film-discover">
							<DropdownItem><a href="/genres/film">Genres</a></DropdownItem>
							<DropdownItem><a href="/releases/film">Releases</a></DropdownItem>
						</DropdownMenu>
					</Dropdown>
					<DropdownItem divider />
					<Dropdown isOpen={booksOpen} toggle={() => (booksOpen = !booksOpen)} direction={'right'}>
						<DropdownToggle caret nav>Books</DropdownToggle>
						<DropdownMenu id="book-discover">
							<DropdownItem><a href="/genres/books">Genres</a></DropdownItem>
							<DropdownItem><a href="/releases/books">Releases</a></DropdownItem>
							<DropdownItem><a href="/authors">Authors</a></DropdownItem>
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
			<a href="/form/media/add">Add Media</a>
		</span>
		<span class="profile-controls">
			<ProfileControls {nickname} />
		</span>
	{:else}
		<span class="search">
			<Search />
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
