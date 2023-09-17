import { writable } from "svelte/store";
import axios from "axios";
import type { Writable } from "svelte/store";
import type { Cast } from "$lib/types/film_tv";

const initialState: Cast = {
  ID: 0,
  actors: [],
  directors: [],
};

interface CastStore extends Writable<Cast> {
  getCast: (media_id: string) => Promise<void>;
}

function createCastStore(): CastStore {
  const { subscribe, set, update } = writable<Cast>(initialState);

  return {
    subscribe,
    set,
    update,
    getCast: async (media_id: string) => {
      try {
        const response = await axios.get(`/api/media/${media_id}/cast/`, {
          headers: { "Content-Type": "application/json" },
        });

        const responseData = response.data;

        if (!responseData || !responseData.data || !Array.isArray(responseData.data)) {
          throw new Error("No data returned from the server");
        }

        console.debug("castData:", response.data);

        const castData = responseData.data;

        set({
          ...initialState,
          ID: castData.ID,
          actors: castData.actors,
          directors: castData.directors,
        });
      }
      catch (error) {
        console.error("Error in getCast:", error);
      }
    },
  };
}

export const castStore = createCastStore();
