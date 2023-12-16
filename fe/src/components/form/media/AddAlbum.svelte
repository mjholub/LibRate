<script lang="ts">
	import { PlusIcon } from 'svelte-feather-icons';
	import type { Album } from '$lib/types/music';
	let album: Album = {
		UUID: '',
		kind: 'album',
		image_paths: [],
		media_id: '',
		name: '',
		title: '',
		created: new Date(),
		creator: null,
		creators: [],
		added: new Date(),
		album_artists: {
			person_artist: [],
			group_artist: []
		},
		release_date: '',
		genres: [],
		duration: {
			Valid: false,
			Time: '00:00:00'
		},
		tracks: []
	};

	const addMore = () => {};
	const addImage = () => {};
</script>

<input type="file" accept="image/*" on:click={addImage} />
<input bind:value={album.name} />
<input bind:value={album.album_artists} />
<!-- render a dropdown allowing to choose whether to add person_artist or group_artist -->
<button id="add-more" on:click={addMore}>
	<PlusIcon />
</button>
<input bind:value={album.release_date} type="date" />
<input bind:value={album.genres} />
<input bind:value={album.duration} type="time" />
<p>Tracks:</p>
