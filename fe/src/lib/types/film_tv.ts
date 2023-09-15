import type { Person } from './people';
import type { UUID } from './utils';
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
  duration?: number;
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
  CastID: number;
  PersonID: number;
};

export type DirectorCast = {
  CastID: number;
  PersonID: number;
};

export type Cast = {
  ID: number;
  Actors: ActorCast[];
  Directors: DirectorCast[];
}
