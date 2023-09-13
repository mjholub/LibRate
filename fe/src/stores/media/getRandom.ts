import { writable } from 'svelte/store';
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
        const response = await fetch(`/api/media/random`, {
          method: 'GET',
          headers: { 'Content-Type': 'application/json' },
        });

        if (!response.ok) {
          console.error('Response not ok: ', response);
          throw new Error(`Response not ok: ${response.status} ${response.statusText}`);
        }

        const responseBody = await response.json();
        const mediaData: AnyMedia[] = responseBody.data;

        if (!mediaData) {
          console.error('No data returned from the server');
          throw new Error('No data returned from the server');
        }

        console.debug(`mediaData: `, mediaData);

        const mediaTypes = determineMediaTypes(mediaData);
        mediaTypes.forEach((mediaType) => {
          set({
            ...initialRandomState, mediaType, [mediaType.toLowerCase()]: mediaData,
            isLoading: false
          });
        });
      } catch (error) {
        console.error('Error in getRandom: ', error);
        throw error;
      }
    },
  };
}

const determineMediaTypes = (mediaData: AnyMedia[]): Array<'Album' | 'Film' | 'Book' | 'Track' | 'TVShow' | 'Unknown'> => {
  if (mediaData.length === 0) {
    throw new Error('Empty media data');
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
