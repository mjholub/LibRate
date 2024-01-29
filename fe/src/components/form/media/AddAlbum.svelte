<script lang="ts">
	import axios from 'axios';
	import { filterXSS } from 'xss';
	import {
		Card,
		Label,
		Input,
		FormGroup,
		ListGroup,
		ListGroupItem
	} from '@sveltestrap/sveltestrap';
	// @ts-ignore
	// @ts-ignore
	import * as time from 'timedstorage/time';
	import { PlusIcon, XIcon } from 'svelte-feather-icons';
	import type { Album, AlbumArtist, Track } from '$lib/types/music';
	import { getMaxFileSize } from '$stores/form/upload';
	import { genreStore } from '$stores/media/genre';
	import searchQueryStore from '$stores/search/websocket';
	import { onMount, onDestroy } from 'svelte';
	import { openFilePicker } from '$stores/form/upload';
	import type { CustomHttpError } from '$lib/types/error';
	import type { NullableDuration } from '$lib/types/utils';
	// @ts-ignore
	import Tags from 'svelte-tags-input';
	import ErrorModal from '$components/modal/ErrorModal.svelte';
	import Typeahead from 'svelte-typeahead';
	import AddTrack from './AddTrack.svelte';

	const basicGenres = [
		'Blues',
		'Classical',
		'Comedy',
		'Country',
		'Darkwave',
		'Easylistening',
		'Experimental',
		'Fieldrecordings',
		'Gospel',
		'Ambient',
		'Dance',
		'Industrial_noise',
		'Marching',
		'Metal',
		'Musical_theatre',
		'Newage',
		'Psychedelia',
		'Punk',
		'Regional',
		'Rnb',
		'Jazz',
		'Electronic',
		'Folk',
		'Singer_songwriter',
		'Rock',
		'Shibuya-Kei',
		'Ska',
		'Sound_effects',
		'Spokenword',
		'Pop',
		'Hip-Hop'
	];
	const reader = new FileReader();
	export let nickname: string;
	let ws: WebSocket;
	let searchResults: any[] = [];
	let genreNames: string[] = [];
	let usingBasicGenresOnly: boolean = false;
	let JSONfile: FileList;
	let jsonParsed: boolean = false;
	let releaseDate: Date;
	let releaseDateString: string;
	let maxFileSize: number;
	let maxFileSizeString: string;
	let imagePaths: string[] = [];
	let errorMessages: CustomHttpError[] = [];
	let availableImportSources: string[] = [];
	let isUploading: boolean = false;
	let imageBase64 = '';
	let importSource = '';
	let importURL = '';
	let currentGenres: any[] = [];

	let hasImportFinished = false;
	let remoteArtistsNames: string[] = [];
	let isArtistsListAmbiguous = false;
	let artistsToBeResolved: AlbumArtist[] = [];
	let firstArtistName = '';
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
		album_artists: [],
		release_date: '',
		genres: [],
		duration: {
			Valid: false,
			Time: '00:00:00'
		},
		tracks: []
	};

	// check if the user has provided more specific genre names (i.e. subgenres)
	// by comparing album.genres against basicGenres
	const checkGenreGranularity = () => {
		if (!album.genres || album.genres.length === 0 || !album.genres[0].name) {
			console.debug('no genres provided');
			usingBasicGenresOnly = false;
			return;
		}
		console.debug('album genres', album.genres);
		for (let i = 0; i < album.genres.length; i++) {
			if (!basicGenres.includes(album.genres[i].name)) {
				console.debug(`genre ${album.genres[i].name} is not a basic genre`);
				usingBasicGenresOnly = false;
				return;
			}
		}
		usingBasicGenresOnly = true;
		console.debug('only basic genres provided');
	};

	const shouldResetAlbum = () => {
		const albumWithLifeTime = localStorage.getItem('album');
		if (albumWithLifeTime) {
			const albumWithLifeTimeObject = JSON.parse(albumWithLifeTime);
			const albumLifeTime = albumWithLifeTimeObject.timeStamp;
			const currentTime = new Date().getTime();
			const timeDifference = currentTime - albumLifeTime;
			const timeDifferenceInHours = timeDifference / (1000 * 3600);
			if (timeDifferenceInHours > 1) {
				return true;
			}
		}
		return false;
	};

	onMount(async () => {
		const importSources = await fetch('/api/media/import-sources').then((res) => res.json());
		availableImportSources = importSources;
		availableImportSources = [...availableImportSources];
		await loadGenreNames();
		maxFileSize = await getMaxFileSize();
		if (shouldResetAlbum()) {
			localStorage.removeItem('album');
		}
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
		// save album, clear after 1 hour
		addEventListener('beforeunload', () => {
			let albumWithLifeTime: Object = {
				timeStamp: new Date().getTime(),
				data: album
			};
			localStorage.setItem('album', JSON.stringify(albumWithLifeTime));
		});
		ws = searchQueryStore.createWebSocket(window.location.host);
		ws.onmessage = (e) => {
			const data = JSON.parse(e.data);
			searchResults = data.results;
		};
	});

	const loadGenreNames = async () => {
		// reassign to avoid 'genreNames is undefined'
		genreNames = [];
		try {
			if (localStorage.getItem('genreNames')) {
				const timestamp = localStorage.getItem(`genreNames_timestamp`);
				const lastTS = parseInt(timestamp || '0', 10);
				const now = new Date().getTime();

				if (now - lastTS > time.DAY) {
					await fetchGenreNames('older than 1 day');
				} else {
					console.debug('getting genre names from local storage');
					genreNames = JSON.parse(localStorage.getItem('genreNames') || '');
					genreNames = [...genreNames];
					if (genreNames.length === 0) {
						await fetchGenreNames('empty');
					}
				}
			} else {
				await fetchGenreNames('not in local storage');
			}
		} catch (error) {
			console.error(error);
			errorMessages.push({
				message: 'Error getting genre links',
				status: 500
			});
		}
	};

	const fetchGenreNames = async (message: string) => {
		console.debug(`genre names are ${message}, getting genre names from API endpoint`);
		const genresResponse = await genreStore.getGenreNames('music', false);
		genreNames.push(...genresResponse);
		genreNames = [...genreNames];
		localStorage.setItem('genreNames', JSON.stringify(genreNames));
		localStorage.setItem(`genreNames_timestamp`, new Date().getTime().toString());
	};

	const handleFileLoad = (e: ProgressEvent<FileReader>) => {
		try {
			const jsonData = JSON.parse(e.target?.result as string);
			const filteredJSON = filterXSS(jsonData);
			releaseDate = parseJSONDate(jsonData.release_date) || new Date();
			Object.assign(album, filteredJSON);
			album.duration = sumAlbumDuration(album.tracks);
			// Trigger a reassignment to make Svelte detect the changes
			album = { ...album };
			album.release_date = releaseDate;
			jsonParsed = true;
		} catch (err) {
			console.error(err);
		}
	};

	onDestroy(() => {
		console.debug('calling destroy hooks for reader');
		removeEventListener('load', () => {
			const dropArea = document.querySelector('.drop-area');

			dropArea?.removeEventListener('dragover', (e) => {
				e.preventDefault();
				dropArea.classList.add('highlight');
			});
		});
		reader.onload = null;
		reader.abort();
		maxFileSize = 0;
		isUploading = false;
		availableImportSources = [];
	});

	const addMore = () => {
		// add another search field for album artists
		document
			.querySelector('#album-artists-search')
			?.insertAdjacentHTML(
				'afterend',
				`<input id="album-artists-search" bind:value={album.album_artists} on:input={searchArtists} />`
			);
	};

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

			onDestroy(() => {
				fileReader.onload = null;
				fileReader.onerror = null;
				fileReader.abort();
			});
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

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
	};

	const genreLinkFromName = (genreName: string): string => {
		const genreURI = genreName.toLowerCase().replace(' ', '-');
		const genreLink = window.location.origin + '/genres/music/' + genreURI;
		return genreLink;
	};

	const openGenreLink = (genreName: string) => window.open(genreLinkFromName(genreName), '_blank');

	const importFromWebSource = async (importSource: string, importURL: string) => {
		const csrfToken = document.cookie
			.split('; ')
			.find((row) => row.startsWith('csrf_'))
			?.split('=')[1];
		const res = await fetch('/api/media/import', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-CSRF-Token': csrfToken || ''
			},
			body: JSON.stringify({
				name: importSource,
				uri: importURL
			})
		});

		if (res.status !== 200) {
			errorMessages.push({
				message: 'Error importing album',
				status: res.status
			});
			console.error(errorMessages);
			errorMessages = [...errorMessages];
		}

		const albumData = await res.json();
		if ('album' in albumData && !('remote_artists' in albumData)) {
			album = albumData.album;
			artistsToBeResolved = albumData.artists;
			isArtistsListAmbiguous = true;
		} else if ('album' in albumData && 'remote_artists' in albumData) {
			remoteArtistsNames = albumData.remote_artists;
			album = albumData.album;
		} else {
			album = albumData;
		}
		hasImportFinished = true;
	};

	const parseJSONDate = (dateString: string): Date | null => {
		const dateParts = dateString.split('-');
		if (dateParts.length === 3) {
			const [day, month, year] = dateParts.map(Number);
			return new Date(year, month - 1, day);
		}
		return null;
	};

	const sumAlbumDuration = (tracks: Track[]): NullableDuration => {
		let sum = 0;
		tracks.forEach((track) => {
			sum += track.duration as number;
		});
		return {
			Valid: true,
			// I hate this
			Time: sum as unknown as string
		};
	};

	const bufferSpotifyAlbumImage = async (imageUrl: string) => {
		const res = await fetch(imageUrl);
		const blob = await res.blob();
		const file = new File([blob], 'cover.jpg', { type: 'image/jpeg' });

		imagePaths.push(URL.createObjectURL(file));
		imagePaths = [...imagePaths];
		await updateImageBase64(file);
	};

	const enterImport = (e: KeyboardEvent) => {
		if (e.key === 'Enter') {
			e.preventDefault();
			importFromWebSource(importSource, importURL);
		} else {
			null;
		}
	};

	const searchArtists = async (e: Event) => {
		if (e.target) {
			searchQueryStore.performSearch((e.target as HTMLInputElement).value, ws);
		}
	};

	// timeout of 60 seconds to load genres
	const genresLoaded = async () => {
		const timeout = 60000;
		const start = new Date().getTime();

		while (new Date().getTime() - start < timeout && genreNames.length === 0) {
			await new Promise((resolve) => setTimeout(resolve, 1000));
		}
		if (genreNames.length === 0) {
			errorMessages.push({
				message: 'Timeout while loading genres',
				status: 500
			});
		}
		errorMessages = [...errorMessages];
	};

	const writeGenres = async () => {
		console.debug('writing genres: ', currentGenres);
		album.genres = [];
		currentGenres.forEach(async (genreName) => {
			const genre = await genreStore.getGenre('music', 'en', genreName);
			if (genre && genre.name !== '') {
				album.genres!.push(genre);
				album.genres = [...album.genres!];
			} else {
				errorMessages.push({
					message: `Genre ${genreName} not found`,
					status: 404
				});
				errorMessages = [...errorMessages];
			}
		});
	};

	$: {
		maxFileSizeString = `${(maxFileSize / 1024 / 1024).toFixed(2)} MB`;

		imagePaths = album.image_paths || [];
		album = { ...album };

		if (album.genres) {
			checkGenreGranularity();
		}
		if (JSONfile && !jsonParsed) {
			reader.onload = handleFileLoad;
			reader.readAsText(JSONfile[0]);
		}

		if (currentGenres.length > 0) {
			async () => {
				await writeGenres();
				checkGenreGranularity();
			};
		}

		isUploading = imagePaths.length !== 0;
		releaseDateString = releaseDate ? releaseDate.toISOString().split('T')[0] : '';
	}
</script>

<svelte:head>
	<link
		rel="stylesheet"
		href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css"
	/>
</svelte:head>

<div class="import-selector">
	<FormGroup>
		<Label id="import-label">Import from:</Label>
		<Input type="select" id="import-source" bind:value={importSource}>
			<option value="">Select a source</option>
			{#each availableImportSources as source}
				{#if source == 'rym'}
					<option value={source}>RateYourMusic</option>
				{:else if source == 'mediawiki'}
					<option value={source}>Wiki</option>
				{:else if source == 'id3'}
					<option value={source}>Music file (ID3 Tags)</option>
				{:else if source == 'json'}
					<option value={source}>JSON</option>
				{:else}
					<option value={source}>{source.charAt(0).toUpperCase() + source.slice(1)}</option>
				{/if}
			{/each}
		</Input>
		{#if importSource != '' && importSource != 'id3' && importSource != 'json'}
			{#if importSource == 'spotify'}
				<p aria-labelledby="spotify-info">Spotify album URL:</p>
			{:else if importSource == 'rym'}
				<p aria-labelledby="rym-info">RateYourMusic album URL:</p>
			{:else if importSource == 'discogs'}
				<p aria-labelledby="discogs-info">Discogs album URL:</p>
			{:else if importSource == 'lastfm'}
				<p aria-labelledby="lastfm-info">Last.fm album URL:</p>
			{:else if importSource == 'listenbrainz'}
				<p aria-labelledby="listenbrainz-info">ListenBrainz album URL:</p>
			{:else if importSource == 'bandcamp'}
				<p aria-labelledby="bandcamp-info">Bandcamp album URL:</p>
			{:else if importSource == 'mediawiki'}
				<p aria-labelledby="mediawiki-info">Wiki album URL:</p>
			{:else if importSource == 'pitchfork'}
				<p aria-labelledby="pitchfork-info">Pitchfork album URL:</p>
			{/if}
			<Input
				type="text"
				id="import-url"
				bind:value={importURL}
				on:keydown={enterImport}
				tabindex="0"
				role="input"
			/>
		{:else if importSource == 'json'}
			<p aria-labelledby="json-info">
				JSON (see <a href="https://codeberg.org/mjh/LibRate/wiki/Album-JSON-fields">specification</a
				>):
			</p>
			<input
				accept="application/json"
				bind:files={JSONfile}
				id="avatar"
				name="avatar"
				type="file"
			/>
		{:else if importSource == 'id3'}
			<p aria-labelledby="id3-info">Music file (ID3 Tags, only MP3 supported):</p>
			<Input type="file" id="import-file" accept="audio/mpeg" bind:value={JSONfile} />
		{/if}
	</FormGroup>
</div>

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
<div class="input-field-element">
	<label for="name">Album Name:</label>
	<input id="name" bind:value={album.name} />
</div>

<div class="input-field-element">
	<label for="album-artists">Album Artists:</label>
	{#if album.album_artists.length === 0}
		<Typeahead
			placeholder="Search for artists"
			on:input={searchArtists}
			{searchResults}
			bind:value={firstArtistName}
			extract="{(artist) => artist.name})}"
			on:clear={() => (album.album_artists = [])}
		/>
	{:else}
		{#each album.album_artists as artist, index}
			<input
				bind:value={album.album_artists[index].name}
				id={`artistName${index}`}
				name={`artistName${index}`}
			/>
		{/each}
	{/if}
</div>
{#if remoteArtistsNames.length > 0 && hasImportFinished}
	<p>The following artists were found in the import source, but not in the database:</p>
	<ListGroup>
		{#each remoteArtistsNames as artistName}
			<ListGroupItem>{artistName} [<a href="form/artist/add" target="_blank">Add</a>]</ListGroupItem
			>
		{/each}
	</ListGroup>
{/if}
{#if isArtistsListAmbiguous}
	<p>More than one artist was found for this album. Select one that matches.</p>
	<div class="artist-selection-grid">
		{#each artistsToBeResolved as artist}
			<Card class="artist-card">
				bind:value={album.album_artists}
				on:click={() => {
					album.album_artists = [artist];
					artistsToBeResolved = [];
					isArtistsListAmbiguous = false;
				}}
				on:keydown={(e) => {
					if (e.key === 'Enter') {
						album.album_artists = [artist];
						artistsToBeResolved = [];
						isArtistsListAmbiguous = false;
					}
				}}
				tabindex="0" role="button"
				<p><a href="/artists/${artist.artist}" target="_blank">{artist.name}</a></p>
			</Card>
		{/each}
	</div>
{/if}

<button id="add-more" on:click={addMore}>
	<PlusIcon />
</button>

<div class="input-field-element">
	<label for="release-date">Release Date:</label>
	<input id="release-date" bind:value={releaseDateString} type="date" />
</div>

<div class="input-field-element">
	<label for="genres">Genres (comma separated):</label>
	{#await genresLoaded()}
		<div class="spinner" />
	{:then}
		<div class="genre-box">
			<Tags
				bind:tags={currentGenres}
				on:tagsChange={console.log(currentGenres)}
				on:focus={console.log(currentGenres)}
				onlyUnique={true}
				autoComplete={localStorage.getItem('genreNames')
					? JSON.parse(localStorage.getItem('genreNames') || '')
					: []}
				onTagClick={openGenreLink}
				onlyAutocomplete={true}
			/>
		</div>
		{#if currentGenres.length > 0}
			{#await writeGenres()}
				<div class="spinner" />
			{:then}
				<p class="success-box">Changes saved!</p>
			{:catch error}
				<p>{error.message}</p>
			{/await}
		{/if}
	{:catch error}
		<p>{error.message}</p>
	{/await}
</div>
{#if usingBasicGenresOnly}
	<div class="notice-box">
		<!-- TODO: if the artist is known, suggest genres based on the genres of previous releases -->
		<p>
			Only basic genres were provided. To find out which subgenres of the provided, try clicking on
			one of the provided genre names to open a new tab with the genre page. There you can find a
			list of subgenres.
		</p>
		<button id="dismiss-notice" on:click={() => (usingBasicGenresOnly = false)}>
			<XIcon />
		</button>
	</div>
{/if}

<div class="input-field-element">
	<label for="duration"
		>Duration (will be calculated automatically from total duration of tracks):</label
	>
	<input id="duration" bind:value={album.duration.Time} type="time" />
</div>

<p>Tracks:</p>
<AddTrack receivedAlbumID={album.UUID} />

<button on:click={handleSubmit}>Submit</button>

{#if errorMessages.length > 0}
	<ErrorModal showErrorModal={true} {errorMessages} />
{/if}

<style>
	.input-field-element {
		margin-bottom: 1rem;
		display: grid;
	}

	button#add-more {
		display: inline-block;
		position: relative;
	}

	button#dismiss-notice {
		top: -0.33em;
		position: relative;
		right: 0.25em;
		padding: 0.2em 0.05em 0.05em;
		border-radius: var(--button-border-radius);
	}

	input#album-artists-search {
		max-width: 90%;
		display: inline-block;
	}

	label {
		display: block;
	}

	.notice-box {
		background-color: #f9ffc4 !important;
		color: #000 !important;
		border: 1px solid #faf0ff;
		border-radius: 4px;
		font-weight: bold;
		font-size: 1rem;
		opacity: 0.85;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

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
