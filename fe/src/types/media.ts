import type { Person } from './people';
import type { UUID } from './utils';

export type Media = {
  UUID: string;
  kind: string;
  name: string;
  genres: Genre[];
  keywords: string[];
  lang_ids: number[];
  creators: Person[];
}

export type Genre = {
  id: number;
  media_id: string; // UUID
  name: string;
  desc_short: string;
  desc_long: string;
  keywords: string[]
  parent_genre: Genre;
  children: Genre[];
}

export type Keyword = {
  id: number;
  keyword: string;
  media_id: UUID;
  stars: number;
}
