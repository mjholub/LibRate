import { writable } from 'svelte/store';
import { initialPerson } from './people.ts';
import type { Writable } from 'svelte/store';
import type { UUID } from '$lib/types/utils.ts';
import type { Person } from '$lib/types/people.ts';
import type { Book } from '$lib/types/books.ts';
import type { Album, Track } from '$lib/types/music.ts';
import type { Film, TVShow } from '$lib/types/film_tv.ts';


export interface MediaStoreState {
  media_id: UUID | UUID[] | null;
  mediaType: 'Album' | 'Film' | 'TVShow' | 'Book' | 'Track' | 'Unknown' | null;
  isLoading: boolean;
  mediaTitle?: string;
  mediaKind?: string;
  created?: Date;
  mediaCreator?: Person;
  album?: Album | Album[] | null;
  book?: Book | Book[] | null;
  track?: Track | Track[] | null;
  film?: Film | Film[] | null;
  tvShow?: TVShow | TVShow[] | null;
};

interface MediaStore extends Writable<MediaStoreState> {
  getMedia: (mediaID: UUID) => Promise<void>;
};

const initialState: MediaStoreState = {
  media_id: null,
  mediaType: null,
  isLoading: true,
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
      console.log(media);
      update((state: MediaStoreState) => ({ ...state, ...media }));
    },
  };
}

export const mediaStore: MediaStore = createMediaStore();
