import { writable } from 'svelte/store';
import { initialPerson } from './people.ts';
import type { Writable } from 'svelte/store';
import type { UUID } from '../../types/utils.ts';
import type { Person } from '../../types/people.ts';


export interface MediaStoreState {
  mediaID: UUID | UUID[] | null;
  mediaTitle?: string;
  mediaKind?: string;
  created?: Date;
  mediaCreator?: Person;
};

interface MediaStore extends Writable<MediaStoreState> {
  getMedia: (mediaID: UUID) => Promise<void>;
};

const initialState: MediaStoreState = {
  mediaID: null,
  mediaTitle: '',
  mediaKind: '',
  created: new Date(),
  mediaCreator: { ...initialPerson },
};

function createMediaStore(): MediaStore {
  const { subscribe, set, update } = writable<MediaStoreState>(initialState);

  return {
    subscribe,
    set,
    update,
    getMedia: async (mediaID: UUID) => {
      const response = await fetch(`/api/media/${mediaID}`);
      const media = await response.json();
      update((state: MediaStoreState) => ({ ...state, ...media }));
    },
  };
}

export const mediaStore: MediaStore = createMediaStore();
