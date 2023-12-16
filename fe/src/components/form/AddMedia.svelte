<script lang="ts">
	import { onDestroy } from 'svelte';
	import { submitMediaForm } from '$stores/form/add_media';
	import type { Media } from '$lib/types/media';
	import Search from '$components/utility/Search.svelte';
	import Footer from '$components/footer/footer.svelte';
	import AddAlbum from './media/AddAlbum.svelte';

	let media: Media;

	const unsubscribe = submitMediaForm.subscribe((value) => {
		media = value;
	});

	submitMediaForm.subscribe((value) => {
		media = value;
	});

	onDestroy(() => {
		unsubscribe();
		submitMediaForm.set(media);
	});

	// return value doesn't matter here
	const submitMedia = async (media: Media) => {
		try {
			await submitMediaForm.submitMedia(media);
		} catch (error) {
			console.error(error);
		}
	};

	const handleSubmit = (e: Event) => {
		e.preventDefault();
		submitMedia(media);
	};
</script>

<div class="form-body" />
<Search />
<h2 class="form-title">Add Media</h2>
<form on:submit|preventDefault={handleSubmit} />
<select bind:value={media.kind} />
{#if media.kind === 'album'}
	<AddAlbum />
{/if}
<Footer />
