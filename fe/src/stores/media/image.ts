import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { MediaImage } from "../../types/media";
import type { UUID } from "../../types/utils";

interface MediaImageStoreState {
  mediaID?: UUID;
  images: MediaImage[];
  mainImage: MediaImage;
  mainImagePath?: string;
};

interface MediaImageStore extends Writable<MediaImageStoreState> {
  getImagesByMedia: (mediaID: UUID) => Promise<void>;
  setMainImage: (image: MediaImage) => void;
};

const initialState: MediaImageStoreState = {
  images: [],
  mainImage: {
    mediaID: "",
    imageID: 0,
    isMain: false,
  },
};

function createMediaImageStore() {
  const { set, subscribe, update } = writable<MediaImageStoreState>(initialState);

  return {
    subscribe,
    set,
    update,
    reset: () => set(initialState),

    getImagesByMedia: async (mediaID: UUID) => {
      const response = await fetch(`/api/media/${mediaID}/images`);
      const images = await response.json();
      update((state: MediaImageStoreState) => ({ ...state, images }));
    },

    setMainImage: (image: MediaImage) => update((state: MediaImageStoreState) => {
      state.mainImage = image;
      return state;
    })
  };
}

export const mediaImageStore: MediaImageStore = createMediaImageStore();
