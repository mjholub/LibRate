import { writable } from 'svelte/store';
import axios from 'axios';
import type { Writable } from 'svelte/store';
import type { AnyMedia } from '$lib/types/media.ts';
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

        const processedMediaData = mediaData.map((media: any) => {
          const kind = media.kind;
          const details = media.details;

          return { kind, ...details };
        });

        mediaTypes.forEach((mediaType) => {
          set({
            ...initialRandomState,
            mediaType,
            [mediaType.toLowerCase()]: processedMediaData,
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
    switch (media.kind) {
      case 'album':
        mediaTypes.push('Album');
        break;
      case 'track':
        mediaTypes.push('Track');
        break;
      case 'film':
        mediaTypes.push('Film');
        break;
      case 'tvshow':
        mediaTypes.push('TVShow');
        break;
      default:
        mediaTypes.push('Unknown');
        break;
    }
  });

  return mediaTypes;
}

export const randomStore: RandomStore = createRandomStore();
