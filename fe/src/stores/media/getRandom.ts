import { writable } from 'svelte/store';
import axios from 'axios';
import type { Writable } from 'svelte/store';
import type { AnyMedia, Media } from '$lib/types/media.ts';
import type { MediaStoreState } from './media.ts';

interface RandomStore extends Writable<MediaStoreState> {
  getRandom: () => Promise<void>;
};

const initialRandomState: MediaStoreState = {
  media_id: null,
  mediaType: null,
  isLoading: true,
};

function createRandomStore(): RandomStore {
  const { subscribe, set, update } = writable<MediaStoreState>(initialRandomState);

  return {
    subscribe,
    set,
    update,
    getRandom: async () => {
      try {
        const response = await axios.get('/api/media/random/', {
          headers: { 'Content-Type': 'application/json' },
        });

        const responseData = response.data;

        if (!responseData || !responseData.data || !Array.isArray(responseData.data)) {
          throw new Error('No data returned from the server');
        }

        console.debug('mediaData:', response.data);

        const mediaData = responseData.data;
        const mediaTypes = determineMediaTypes(mediaData);

        // filter media data by type
        //const albums = mediaData.filter((media) => media.media_id && media.album_artists);

        mediaTypes.forEach((mediaType) => {
          set({
            ...initialRandomState,
            mediaType,
            [mediaType.toLowerCase()]: mediaData,
            isLoading: false,
          });
        });
      } catch (error) {
        console.error('Error in getRandom:', error);
      }
    },
  }
}

const determineMediaTypes = (mediaData: AnyMedia[]): Array<'Album' | 'Film' | 'Book' | 'Track' | 'TVShow' | 'Unknown'> => {
  if (!mediaData || mediaData.length === 0) {
    // NOTE: not throwing error to have some null safety
    return ['Unknown'];
  }

  const mediaTypes: Array<'Album' | 'Film' | 'Book' | 'Track' | 'TVShow' | 'Unknown'> = [];

  mediaData.forEach((media) => {
    if ('media_id' in media && 'album_artists' in media) {
      mediaTypes.push('Album');
    } else if ('media_id' in media && 'track_number' in media && 'album_id' in media) {
      mediaTypes.push('Track');
    } else if ('publication_date' in media && 'authors' in media && 'pages' in media) {
      mediaTypes.push('Book');
    } else if (media.kind === 'film') {
      mediaTypes.push('Film');
    } else if (media.kind === 'tvshow') {
      mediaTypes.push('TVShow');
    } else {
      mediaTypes.push('Unknown');
    }
  });

  return mediaTypes;
}

export const randomStore: RandomStore = createRandomStore();
