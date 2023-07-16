import type { Either } from 'typescript-monads'
import type { Person, Studio, Group } from './people';
import type { Genre, Media } from './media';
import type { UUID } from './utils';

export interface Album extends Media {
  media_id: UUID;
  name: string;
  album_artists: Either<Person[], Group[]>;
  image_paths: string[];
  release_date: Date;
  genres?: Genre[];
  studio?: Studio;
  keywords?: string[];
  duration: number;
  tracks: Track[];
  languages?: string[];
}

export interface Track extends Media {
  media_id: UUID;
  track_number: number;
  name: string;
  artists: Either<Person[], Group[]>;
  duration: number;
  lyrics?: string;
  languages?: string[];
}
