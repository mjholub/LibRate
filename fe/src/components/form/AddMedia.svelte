<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { submitMediaForm } from '../../stores/form/add_media';
	import type { Media } from '../../types/media';

	let media: Media;
	submitMediaForm.subscribe((value) => {
		media = value;
	});

	onDestroy(() => {
		submitMediaForm.set(media);
	});

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
<h2 class="form-title">Add Media</h2>
<form on:submit|preventDefault={handleSubmit} />
