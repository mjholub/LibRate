import type { Person, Group } from './people';
import type { UUID } from './utils';
import type { Album, Track } from './music';
import type { Book } from './books';
import type { Film, TVShow } from './film_tv';
import type { Tag } from 'language-tags';

export type AnyMedia = Album | Track | Book | Film | TVShow;

export interface Media {
  UUID: string | UUID;
  kind: string;
  title: string;
  created: Date;
  creator: Person | Group | null; // WARN: in the backend code it only references the ID (nullable int32)
  creators: (Person | Group)[];
  added: Date;
  modified?: Date;
};

export type MediaImage = {
  mediaID: UUID;
  imageID: number;
  isMain: boolean;
};


export type Genre = {
  id: number;
  name: string;
  description: GenreDescription[] | GenreDescription | null;
  keywords: string[]
  parent_genre: number | null;
  children: number[] | null;
};

export type GenreDescription = {
  genre_id: number;
  language: Tag; // IANA tag
  description: string;
}

export type Keyword = {
  id: number;
  keyword: string;
  stars: number;
  vote_count: number;
  avg_score?: number | null;
};

export const lookupGenre = async (name: string): Promise<Genre | null> => {
  return new Promise(async (resolve, reject) => {
    const res = await fetch(`/api/genres/${name}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        },
      });
    switch (res.status) {
      case 200:
        resolve(await res.json());
        break;
      case 404:
        resolve(null);
        break;
      default:
        reject(res.statusText);
        break;
    }
  });
};
