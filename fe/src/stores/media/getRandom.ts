import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { UUID } from '../../types/utils.ts';
import type { Media } from '../../types/media.ts';
import type { MediaStoreState } from './media.ts';

interface RandomStore extends Writable<MediaStoreState> {
  getRandom: () => Promise<void>;
};

const initialRandomState: MediaStoreState = {
  mediaID: null,
  mediaType: null,
};

function createRandomStore(): RandomStore {
  const { subscribe, set, update } = writable<MediaStoreState>(initialRandomState);

  return {
    subscribe,
    set,
    update,
    getRandom: async () => {
      const response = await fetch(`/api/media/random`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
      });

      if (!response.ok) {
        console.error(response.statusText);
        throw new Error("Failed to fetch media");
      }

      const mediaData: Media[] = await response.json();
      console.debug(`mediaData: `, mediaData);

      const mediaType = determineMediaType(mediaData);
      set({ ...initialRandomState, mediaType, [mediaType.toLowerCase()]: mediaData });
    }
  };
};

const determineMediaType = (mediaData: Media[]): 'Album' | 'Film' | 'Book' | 'Track' | 'TVShow' => {
  if (mediaData.length === 0) {
    throw new Error('Empty media data');
  }

  const firstMedia = mediaData[0];

  if ('media_id' in firstMedia && 'album_artists' in firstMedia) {
    return 'Album';
  }

  if ('media_id' in firstMedia && 'track_number' in firstMedia && 'artists' in firstMedia) {
    return 'Track';
  }

  if ('publication_date' in firstMedia && 'authors' in firstMedia && 'pages' in firstMedia) {
    return 'Book';
  }

  if (firstMedia.kind === 'film') {
    return 'Film';
  }

  if (firstMedia.kind === 'tvshow') {
    return 'TVShow';
  }

  throw new Error('Unknown media type');
}

export const randomStore: RandomStore = createRandomStore();
