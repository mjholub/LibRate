import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Genre } from '$lib/types/media.ts';
import axios from 'axios';

interface GenreStoreState {
  genres: Genre[];
};

type column = 'name' | 'id' | 'kinds' | 'parent' | 'children' | null
type kind = 'music' | 'film' | 'tv' | 'book' | 'game' | null

interface GenreStore extends Writable<GenreStoreState> {
  getGenres: (all: boolean, columns: column[], kind: kind) => Promise<void>;
  getGenreNames: (kind: kind, asLinks: boolean) => Promise<string[]>;
  // prefferably IANA code, but might be just the base part, like 'en' or 'de'
  getGenre: (kind: kind, lang: string, genre: string) => Promise<Genre | null>;
};

function createGenreStore(): GenreStore {
  const { subscribe, set, update } = writable<GenreStoreState>({ genres: [] });

  return {
    subscribe,
    set,
    update,
    getGenres: async (all: boolean, columns: column[], kind: kind) => {
      await axios.get(`/api/media/genres/${kind}?all=${all}&columns=${columns.join('&columns=')}`, {
      }).then(response =>
        // NOTE: This is not a error. As ugly it might look, data is actually nested like this.
        set({ genres: response.data.data })
      ).catch(err => {
        console.log(err);
        return [];
      });
    },
    getGenreNames: async (kind: kind, asLinks: boolean) => {
      return new Promise(async (resolve, reject) => {
        await axios.get(`/api/media/genres/${kind}?names_only=true?as_links=${asLinks}`, {
        }).then(res => {
          resolve(res.data.data);
        }).catch(err => {
          console.log(err);
          reject(err);
        });
      });
    },
    getGenre: async (kind: kind, lang: string, genre: string) => {
      return new Promise(async (resolve, reject) => {
        await axios.get(`/api/media/genres/${kind}/${genre}?lang=${lang.toString()}`, {
          params: {
            columns: 'all'
          }
        }).then(res => {
          set({ genres: res.data.data })
          resolve(res.data.data);
        }).catch(err => {
          console.log(err);
          reject(err);
        });
      });
    },
  };
};

export const genreStore: GenreStore = createGenreStore();
