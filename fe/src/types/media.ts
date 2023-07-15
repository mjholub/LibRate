import type { Person } from './people';
import type { UUID } from './utils';

export interface Media {
  UUID: string | UUID;
  kind: string;
  title: string;
  created: Date;
  creator: Person | null;
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
  media_id: UUID;
  stars: number;
};
