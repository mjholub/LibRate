import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Genre } from '$lib/types/media.ts';

interface GenreStoreState {
  genres: Genre[];
};

type column = 'name' | 'id' | 'kinds' | 'parent' | 'children' | null
type kind = 'music' | 'film' | 'tv' | 'book' | 'game' | null

interface GenreStore extends Writable<GenreStoreState> {
  getGenres: (all: boolean, columns: column[], kind: kind) => Promise<void>;
  getGenreNames: (kind: kind, asLinks: boolean) => Promise<string[]>;
  // preferably IANA code, but might be just the base part, like 'en' or 'de'
  getGenre: (kind: kind, lang: string, genre: string) => Promise<Genre | null>;
};

function createGenreStore(): GenreStore {
  const { subscribe, set, update } = writable<GenreStoreState>({ genres: [] });

  return {
    subscribe,
    set,
    update,
    getGenres: async (all: boolean, columns: column[], kind: kind): Promise<void> => {
      try {
        const response = await fetch(`/api/media/genres/${kind}?all=${all}&columns=${columns.join('&columns=')}`);
        const data = await response.json();
        set({ genres: data });
      } catch (err) {
        console.log(err);
        throw err;
      }
    },

    getGenreNames: async (kind, asLinks) => {
      try {
        const response = await fetch(`/api/media/genres/${kind}?names_only=true&as_links=${asLinks}&all=true`);
        const data = await response.json();
        return data;
      } catch (err) {
        console.log(err);
        throw err;
      }
    },

    getGenre: async (kind, lang, genre) => {
      try {
        // format to snake_case, lowercase
        genre = genre.toLowerCase().replace(/ /g, '_');
        const response = await fetch(`/api/media/genre/${kind}/${genre}/?lang=${lang.toString()}`);
        const data = await response.json();
        set({ genres: data.data });
        return data.data;
      } catch (err) {
        console.log(err);
        throw err;
      }
    }
  };
};

export const genreStore: GenreStore = createGenreStore();
