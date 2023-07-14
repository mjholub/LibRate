import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Image } from "../../types/cdn";


interface ImageStoreState {
  images: Image[];
};

interface ImageStore extends Writable<ImageStoreState> {
  getPaths: (id: number) => Promise<void>,
  getAlt: (id: number) => Promise<string>,
};

const initialState: ImageStoreState = {
  images: [],
};

function createImageStore() {
  const { set, subscribe, update } = writable<ImageStoreState>(initialState);

  return {
    subscribe,
    getPaths: async (id: number) => {
      const response = await fetch(`/api/cdn/images/${id}`);
      const images = await response.json();
      set({ images });
    },

    getAlt: async (id: number) => {
      const response = await fetch(`/api/cdn/images/${id}/alt`);
      const alt = await response.text();
      return alt;
    },
  };
};

export const imageStore: ImageStore = createImageStore();

