import type { Person } from './people';
import type { UUID, NullableDuration, NullableString, NullableDate } from './utils';
import type { Media } from './media';

export interface Film {
  title: string;
  castID: number;
  synopsis?: NullableString;
  releaseDate?: NullableDate;
  duration?: NullableDuration | null;
  rating?: number;
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
