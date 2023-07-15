import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { UUID } from '../../types/utils.ts';
import type { MediaStoreState } from './media.ts';

interface RandomStore extends Writable<MediaStoreState> {
  getRandom: () => Promise<void>;
};

const initialRandomState: MediaStoreState = {
  mediaID: null,
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

      const mediaID: UUID[] = await response.json();
      console.debug(`mediaID: ${mediaID}`);
      set({ ...initialRandomState, mediaID });
    }
  };
};

export const randomStore: RandomStore = createRandomStore();
