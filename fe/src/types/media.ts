import type { Person } from './people';
import type { UUID } from './utils';
import type { Film, TVShow } from './film_tv';

export type Media = {
  UUID: string;
  kind: string;
  title: string;
  created: Date;
  creator: Person;
} | Film | TVShow;

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
  media_id: UUID;
  stars: number;
};
