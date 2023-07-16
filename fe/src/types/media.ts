import type { Person, Group } from './people';
import type { UUID } from './utils';
import type { Album, Track } from './music';
import type { Book } from './books';
import type { Film, TVShow } from './film_tv';

export type AnyMedia = Album | Track | Book | Film | TVShow;

export interface Media {
  UUID: string | UUID;
  kind: string;
  title: string;
  created: Date;
  creator: Person | Group | null;
};

export type MediaImage = {
  mediaID: UUID;
  imageID: number;
  isMain: boolean;
};


export type Genre = {
  id: number;
  name: string;
  desc_short: string;
  desc_long: string;
  keywords: string[]
  parent_genre: Genre;
  children: Genre[];
};

export type Keyword = {
  id: number;
  keyword: string;
  stars: number;
  vote_count: number;
  avg_score?: number | null;
};
