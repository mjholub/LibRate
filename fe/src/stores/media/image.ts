import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { MediaImage } from "$lib/types/media";
import type { UUID } from "$lib/types/utils";

interface MediaImageStoreState {
  mediaID?: UUID;
  //images: MediaImage;
  mainImage: MediaImage;
  mainImagePath?: string;
};

interface MediaImageStore extends Writable<MediaImageStoreState> {
  getImageByMedia: (media_id: UUID) => Promise<void>;
  setMainImage: (image: MediaImage) => void;
};

const initialState: MediaImageStoreState = {
  mainImage: {
    mediaID: "",
    imageID: 0,
    isMain: false,
  },
  mainImagePath: "",
};

function createMediaImageStore() {
  const { set, subscribe, update } = writable<MediaImageStoreState>(initialState);

  return {
    subscribe,
    set,
    update,
    reset: () => set(initialState),

    getImageByMedia: async (media_id: UUID) => {
      const response = await fetch(`/api/media/${media_id}/images`);
      const image = await response.text();
      update((state: MediaImageStoreState) => ({ ...state, mainImagePath: image }));
    },

    setMainImage: (image: MediaImage) => update((state: MediaImageStoreState) => {
      state.mainImage = image;
      return state;
    })
  };
}

export const mediaImageStore: MediaImageStore = createMediaImageStore();
