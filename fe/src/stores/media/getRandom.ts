import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { UUID } from '../../types/utils.ts';

interface RandomStoreState {
  mediaID: UUID[] | null;
};

interface RandomStore extends Writable<RandomStoreState> {
  getRandom: () => Promise<void>;
};

const initialRandomState: RandomStoreState = {
  mediaID: null,
};

function createRandomStore(): RandomStore {
  const { subscribe, set, update } = writable<RandomStoreState>(initialRandomState);
  return {
    subscribe,
    set,
    update,
    getRandom: async () => {
      const response = await fetch(`/api/media/random`,
        {
          method: 'GET',
          headers: { 'Content-Type': 'application/json' },
        }
      );
      response.ok || console.error(response.statusText);
      const mediaID: UUID[] = await response.json();
      console.debug(`mediaID: ${mediaID}`);
      set({ ...initialRandomState, mediaID });
    },
  };
};

export const randomStore: RandomStore = createRandomStore();
