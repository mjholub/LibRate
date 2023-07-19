import type { Person } from './people';
import type { UUID } from './utils';
import type { Media } from './media';

// TODO: add genre, keywords, etc.
// NOTE: these are (to be?) stored in junction tables in the database
// like film_genres, film_keywords etc.
export interface Film extends Media {
  UUID: UUID;
  kind: 'film';
  title: string;
  created: Date;
  creator: Person;
}

export interface TVShow extends Media {
  UUID: UUID;
  kind: 'tvshow';
  title: string;
  created: Date;
  creator: Person;
}
