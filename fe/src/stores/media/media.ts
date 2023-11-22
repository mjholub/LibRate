import { writable } from 'svelte/store';
import { initialPerson } from './people.ts';
import type { Writable } from 'svelte/store';
import type { UUID } from '$lib/types/utils.ts';
import type { Person } from '$lib/types/people.ts';
import type { AnyMedia } from '$lib/types/media.ts';


export interface MediaStoreState {
  media_id: UUID | UUID[] | null;
  mediaType: 'Album' | 'Film' | 'TVShow' | 'Book' | 'Track' | 'Unknown' | null;
  isLoading: boolean;
  mediaTitle?: string;
  created?: Date;
  mediaCreator?: Person;
  media: AnyMedia | null | AnyMedia[];
};

interface MediaStore extends Writable<MediaStoreState> {
  getMedia: (mediaID: UUID) => Promise<void>;
};

const initialState: MediaStoreState = {
  media_id: null,
  mediaType: null,
  isLoading: true,
  mediaTitle: '',
  created: new Date(),
  mediaCreator: { ...initialPerson },
  media: null,
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
      console.log(media);
      update((state: MediaStoreState) => ({ ...state, ...media }));
    },
  };
}

export const mediaStore: MediaStore = createMediaStore();
