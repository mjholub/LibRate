import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Genre } from '$lib/types/media.ts';
import type { Tag } from 'language-tags';
import axios from 'axios';

interface GenreStoreState {
  genres: Genre[];
};

type column = 'name' | 'id' | 'kinds' | 'parent' | 'children' | null
type kind = 'music' | 'film' | 'tv' | 'book' | 'game' | null

interface GenreStore extends Writable<GenreStoreState> {
  getGenres: (all: boolean, columns: column[], kind: kind) => Promise<void>;
  getGenre: (kind: kind, lang: Tag, genre: string) => Promise<Genre | null>;
};

function createGenreStore(): GenreStore {
  const { subscribe, set, update } = writable<GenreStoreState>({ genres: [] });

  return {
    subscribe,
    set,
    update,
    getGenres: async (all: boolean, columns: column[], kind: kind) => {
      const response = await axios.get(`/api/media/genres/${kind}?all=${all}`, {
        params: {
          columns: columns.join(',')
        }
      }).then(res =>
        set({ genres: res.data })
      ).catch(err => {
        console.log(err);
        return [];
      });
    },
    getGenre: async (kind: kind, lang: Tag, genre: string) => {
      return new Promise(async (resolve, reject) => {
        const response = await axios.get(`/api/media/genres/${kind}/${genre}?lang=${lang.toString()}`, {
          params: {
            columns: 'all'
          }
        }).then(res => {
          set({ genres: res.data })
          resolve(res.data);
        }).catch(err => {
          console.log(err);
          reject(err);
        });
      });
    },
  };
};

export const genreStore: GenreStore = createGenreStore();
