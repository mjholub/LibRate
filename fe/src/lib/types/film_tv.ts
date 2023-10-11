import type { Person } from './people';
import type { UUID, NullableDuration } from './utils';
import type { Media } from './media';

// TODO: add genre, keywords, etc.
// NOTE: these are (to be?) stored in junction tables in the database
// like film_genres, film_keywords etc.
export interface Film extends Media {
  media_id: UUID;
  kind: 'film';
  title: string;
  castID: number;
  synopsis?: string;
  releaseDate?: Date;
  duration?: NullableDuration | null;
  rating?: number;
  created: Date;
  creator: Person | null;
}

export interface TVShow extends Media {
  media_id: UUID;
  kind: 'tvshow';
  title: string;
  created: Date;
  creator: Person;
}

export type ActorCast = {
  castID: number;
  personID: number;
};

export type DirectorCast = {
  castID: number;
  personID: number;
};

export type Cast = {
  ID: number;
  actors: ActorCast[];
  directors: DirectorCast[];
}
