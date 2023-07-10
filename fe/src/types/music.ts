import type { Either } from 'typescript-monads'
import type { Person, Studio, Group } from './people';
import type { Genre } from './media';
import type { UUID } from './utils';

export interface Album {
  media_id: UUID;
  name: string;
  album_artists: Either<Person[], Group[]>;
  release_date: Date;
  genres?: Genre[];
  studio?: Studio;
  keywords?: string[];
  duration: number;
  tracks: Track[];
  languages?: string[];
}

export interface Track {
  media_id: UUID;
  name: string;
  artists: Either<Person[], Group[]>;
  duration: number;
  lyrics?: string;
  languages?: string[];
}
