import type { Either } from 'typescript-monads'
import type { Person, Group } from './people';
import type { UUID } from './utils';

export interface Media {
  UUID: string | UUID;
  kind: string;
  title: string;
  created: Date;
  // TODO: this doesn't precisely match the db
  // We need to add a migration that'd allow the creator
  // to not only be a person, but also a group
  creator: Either<Person, Group>;
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
