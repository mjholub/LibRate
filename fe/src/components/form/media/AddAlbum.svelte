<script lang="ts">
	import axios from 'axios';
	import { PlusIcon, XIcon } from 'svelte-feather-icons';
	import type { Album } from '$lib/types/music';
	import { lookupGenre } from '$lib/types/media';
	import { getMaxFileSize } from '$stores/form/upload';
	import { onMount, onDestroy } from 'svelte';
	import { openFilePicker } from '$stores/form/upload';
	import type { CustomHttpError } from '$lib/types/error';
	import AddTrack from './AddTrack.svelte';

	export let nickname: string;
	let maxFileSize: number;
	let maxFileSizeString: string;
	let imagePaths: string[] = [];
	let errorMessages: CustomHttpError[] = [];
	let genreLinks: string[] = [];
	let isUploading: boolean = false;
	let imageBase64 = '';

	onMount(async () => {
		maxFileSize = await getMaxFileSize();
		addEventListener('load', () => {
			const dropArea = document.querySelector('.drop-area');
			if (dropArea) {
				dropArea.addEventListener('dragover', (e) => {
					e.preventDefault();
					dropArea.classList.add('highlight');
				});
				dropArea.addEventListener('dragleave', () => {
					dropArea.classList.remove('highlight');
				});
				dropArea.addEventListener('drop', async (e) => {
					e.preventDefault();
					dropArea.classList.remove('highlight');
					try {
						await updateImage(e);
					} catch (error) {
						errorMessages.push({
							message: "Couldn't upload image",
							status: 500
						});
						errorMessages = [...errorMessages];
					}
				});
			}
		});
	});

	onDestroy(() => {
		maxFileSize = 0;
		isUploading = false;
	});

	// call lookupGenre to check if genre exists in db
	// then split the genres into an array of links to each genre's page
	// If a genre is found, album's genres are updated with the genre's info
	const listGenres = async (genreNames: string[]) => {
		let genreLinks: string[] = [];
		genreNames.forEach(async (genreName) => {
			const genre = await lookupGenre(genreName);
			if (genre) {
				album.genres?.push(genre);
				album.genres = [...(album.genres || [])];
				genreLinks.push(`<a href="/genre/${genre.id}">${genre.name}</a>`);
				genreLinks = [...genreLinks];
			} else {
				errorMessages.push({
					message: `Genre ${genreName} not found`,
					status: 404
				});
				errorMessages = [...errorMessages];
			}
		});
		return genreLinks;
	};

	$: {
		maxFileSizeString = `${(maxFileSize / 1024 / 1024).toFixed(2)} MB`;
		imagePaths = album.image_paths || [];
		album.genres = album.genres || [];
		genreLinks = [];
		isUploading = imagePaths.length !== 0;
	}

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

	// updateImage reactively changes the displayed image when the user uploads a new one
	const updateImage = async (e: Event) => {
		const files = (e.target as HTMLInputElement).files;
		if (files) {
			const f = Array.from(files);
			f.forEach(async (file: File | Blob) => {
				if (file.size > maxFileSize) {
					errorMessages.push({
						message: `File size must be less than ${maxFileSizeString}`,
						status: 413
					});
					errorMessages = [...errorMessages];
				} else {
					imagePaths.push(URL.createObjectURL(file));
					imagePaths = [...imagePaths];
					await updateImageBase64(file);
				}
			});
		}
	};

	const updateImageBase64 = async (file: File | Blob): Promise<void> => {
		return new Promise((resolve, reject) => {
			const fileReader: FileReader = new FileReader();
			fileReader.onload = async () => {
				try {
					const imageBase64WithPrefix: string = fileReader.result as string;
					imageBase64 = imageBase64WithPrefix.split(',')[1]; // remove prefix
					isUploading = false;
					resolve();
				} catch (err) {
					reject(err);
				}
			};
			fileReader.onerror = (e) => reject(e);
			isUploading = true;
			fileReader.readAsDataURL(file);
		});
	};

	// submitImage sends the image to the server once the form has been filled out
	const submitImage = async (e: Event) => {
		return new Promise((resolve, reject) => {
			isUploading = true;
			console.debug('addImage');
			const files = (e.target as HTMLInputElement).files;
			if (files) {
				console.debug('files', files);
				const file = files[0];
				if (file.size > maxFileSize) {
					isUploading = false;
					reject(new Error(`File size must be less than ${maxFileSizeString}`));
				}
				const reader = new FileReader();
				console.info('file reader initialized');
				let csrfToken: string | undefined;
				csrfToken = document.cookie
					.split('; ')
					.find((row) => row.startsWith('csrf_'))
					?.split('=')[1];

				reader.onload = async (e) => {
					console.debug('reader onload');
					const data = e.target?.result;
					if (data) {
						console.debug('data', data);
						const formData = new FormData();
						console.info('form data initialized');
						formData.append('fileData', file);
						formData.append('imageType', 'album_cover');
						formData.append('member', nickname);
						try {
							const res = await axios.post('/api/upload/image', formData, {
								headers: {
									'Content-Type': 'multipart/form-data',
									Authorization: `Bearer ${localStorage.getItem('jwtToken')}`,
									'X-CSRF-Token': csrfToken || ''
								}
							});
							if (res.status === 200) {
								album.image_paths = [res.data.path];
								imagePaths = [res.data.path];
								isUploading = false;
								resolve(res.data.path);
							} else {
								errorMessages.push({
									message: res.data.message,
									status: res.status
								});
								isUploading = false;
								reject(res.status);
							}
						} catch (error) {
							errorMessages.push({
								message: 'Something went wrong',
								status: 500
							});
							isUploading = false;
							reject(error);
						}
					}
				};
				reader.readAsDataURL(file);
			}
		});
	};

	const removeGenre = (index: number) => {
		if (album.genres) {
			album.genres.splice(index, 1);
		}
	};

	const handleGenreAdd = async (e: Event) => {
		e.preventDefault();
		if (genres) {
			const genreNames = genres.split(',');
			await listGenres(genreNames);
		}
	};

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
	};

	let genres: string = '';
</script>

<!-- svelte-ignore  a11y-no-noninteractive-element-interactions -->
<div
	class="drop-area"
	on:drop={updateImage}
	on:click={() => openFilePicker(updateImage, 'image/*')}
	on:keydown={(e) => (e.key === 'Space' ? openFilePicker(updateImage, 'image/*') : null)}
	on:dragover={(e) => e.preventDefault()}
	aria-dropeffect="copy"
	role="region"
	aria-labelledby="drop-area-label"
>
	<p id="drop-area-label">
		<!-- svelte-ignore a11y-missing-attribute -->
		<a
			on:click={() => openFilePicker(updateImage, 'image/*')}
			on:keydown={(e) => (e.key === 'Enter' ? openFilePicker(updateImage, 'image/*') : null)}
			tabindex="0"
			role="button">Drop or click to add album cover here</a
		>
	</p>

	{#if isUploading}
		<div class="spinner" />
	{/if}
	{#if album.image_paths}
		<img src={album.image_paths[0]} alt="Album Cover" />
	{/if}
</div>

<label for="name">Album Name:</label>
<input id="name" bind:value={album.name} />

<label for="album-artists">Album Artists:</label>
<select id="album-artists" bind:value={album.album_artists}>
	<option value="person_artist">Person</option>
	<option value="group_artist">Group</option>
</select>

<button id="add-more" on:click={addMore}>
	<PlusIcon />
</button>

<label for="release-date">Release Date:</label>
<input id="release-date" bind:value={album.release_date} type="date" />

<label for="genres">Genres (comma separated):</label>
<div>
	{#if genreLinks}
		{#each genreLinks as genre, index}
			<div class="genre-box">
				{genre}
				<span
					class="remove-genre"
					on:click={() => removeGenre(index)}
					on:keyup={(e) => e.key === 'Enter' && removeGenre(index)}
					aria-label="Remove genre"
					role="button"
					tabindex="0"
				>
					<XIcon size="12" />
				</span>
			</div>
		{/each}
	{/if}
</div>
<!-- pass to listGenres -->
<input id="genres" bind:value={genres} on:blur={handleGenreAdd} />

<label for="duration">Duration:</label>
<input id="duration" bind:value={album.duration} type="time" />

<p>Tracks:</p>
<AddTrack />

<button on:click={handleSubmit}>Submit</button>

<style>
	.drop-area {
		border: 2px dashed #ccc;
		padding: 20px;
		text-align: center;
	}

	.drop-area img {
		max-width: 100%;
		max-height: 200px;
		margin-top: 10px;
	}

	.genre-box {
		display: inline-block;
		margin: 0 8px 8px 0;
		padding: 6px 12px;
		background-color: #f0f0f0;
		border: 1px solid #ccc;
		border-radius: 4px;
		position: relative;
	}

	.remove-genre {
		cursor: pointer;
		position: absolute;
		top: 50%;
		right: 8px;
		transform: translateY(-50%);
		color: #888;
	}

	.spinner {
		border: 4px solid rgba(0, 0, 0, 0.1);
		border-top: 4px solid #3498db;
		border-radius: 50%;
		width: 20px;
		height: 20px;
		animation: spin 1s linear infinite;
		margin-left: 10px;
		display: inline-block;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>
