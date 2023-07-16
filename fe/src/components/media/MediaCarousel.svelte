<script lang="ts">
	import { onMount } from 'svelte';
	import { randomStore } from '../../stores/media/getRandom.ts';
	import { mediaImageStore } from '../../stores/media/image.ts';
	import { imageStore } from '../../stores/cdn/imagePath.ts';
	import MediaCard from './MediaCard.svelte';
	import AlbumCard from './AlbumCard.svelte';
	import type { Group, Person, Creator } from '../../types/people.ts';
	import type { Media, MediaImage } from '../../types/media.ts';
	import type { Album } from '../../types/music.ts';

	let media: (Media | Album)[] = [];
  let album: Album = {
    media_id: '',
    name: '',
    album_artists: [],
    image_paths: [],
    release_date: new Date(),
    genres: [],
    keywords: [], 
    duration: 0,
    tracks: []
  };
	let mediaImage: MediaImage = {
		mediaID: '',
		imageID: 0,
		isMain: false
	};
	let mediaImgPath = '';
	let creators: Creator[] = [];

	onMount(() => {
		(async () => {
			await randomStore.getRandom();
		})();

		console.info('mounting MediaCarousel initialized');
		console.info('data from randomStore: ', randomStore);

 let subscriptions: (() => void)[] = [];

    let unsubscribe = randomStore.subscribe((data) => {
      if (
        !data.mediaID ||
        !data.mediaTitle ||
        !data.mediaCreator ||
        !data.created ||
        !data.mediaKind
      ) {
        return;
      }

      const newMedia: Media = {
        UUID: data.mediaID[0],
        title: data.mediaTitle,
        kind: data.mediaKind,
        created: data.created,
        creator: data.mediaCreator
      };

      media = [...media, newMedia];

      processMediaItems(media, subscriptions);
    });

    subscriptions.push(unsubscribe);

    return () => {
      subscriptions.forEach((unsub) => unsub());
    };
  });

  async function processMediaItems(mediaItems: (Media | Album)[], subscriptions: (() => void)[]) {
    for (const mediaItem of mediaItems) {
      console.debug('mediaItem: ', mediaItem);
      await mediaImageStore.getImagesByMedia(mediaItem.UUID);

      let mediaImgStrSub = mediaImageStore.subscribe((data) => {
        if (!data || !data.mediaID || data.images.length === 0) {
          return;
        }
        mediaImage = {
          mediaID: data.mediaID,
          imageID: data.images[0].imageID,
          isMain: data.mainImage.isMain
        };
      });

      subscriptions.push(mediaImgStrSub);

      await imageStore.getPaths(mediaImage.imageID);
      console.debug('imageStore: ', imageStore);

      let imgStoreSub = imageStore.subscribe((data) => {
        if (!data || !data.images || data.images.length === 0) {
          return;
        }
        mediaImgPath = data.images[0].source;
      });

      subscriptions.push(imgStoreSub);

      if (mediaItem.kind === 'album') {
        const album = mediaItem as Album;
        const creatorArray = Array.isArray(album.album_artists)
          ? album.album_artists
          : [album.album_artists];

        for (const creator of creatorArray) {
          if ('first_name' in creator) {
            const newPerson = creator as Person;
            const newCreator: Creator = { id: newPerson.id, name: newPerson.name };
            creators.push(newCreator);
          } else {
            const newGroup = creator as Group;
            const newCreator: Creator = { id: newGroup.id, name: newGroup.name };
            creators.push(newCreator);
          }
        }
      } else {
        const creatorArray = Array.isArray(mediaItem.creator)
          ? mediaItem.creator
          : [mediaItem.creator];

        for (const creator of creatorArray) {
          if ('first_name' in creator) {
            const newPerson = creator as Person;
            const newCreator: Creator = { id: newPerson.id, name: newPerson.name };
            creators.push(newCreator);
          } else {
            const newGroup = creator as Group;
            const newCreator: Creator = { id: newGroup.id, name: newGroup.name };
            creators.push(newCreator);
          }
        }
      }
    }
  }

  const isAlbum = (mediaItem: Media | Album): mediaItem is Album => {
    return mediaItem.kind === 'album';
  };
</script>

<div class="carousel">
  {#if media.length === 0}
    <div>Loading...</div>
  {:else}
    {#each media as mediaItem (mediaItem.UUID)}
      <div class="media-card-wrapper">
        {#if isAlbum(mediaItem)}
          <AlbumCard {album} title={mediaItem.title} image={mediaImgPath} {creators} />
        {:else}
          <MediaCard {mediaItem} title={mediaItem.title} image={mediaImgPath} {creators} />
        {/if}
      </div>
    {/each}
  {/if}
</div>

<style>
	.carousel {
		display: flex;
		overflow-x: scroll;
		height: 100%;
		width: 100%;
	}

	.media-card-wrapper {
		flex: 0 0 auto;
		width: 30%;
		height: 40%;
		padding: 1em;
	}
</style>
