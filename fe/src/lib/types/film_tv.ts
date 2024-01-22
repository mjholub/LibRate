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

// TODO: verify if (de)serialization works as expected
export interface TVShow {
  title: string;
  seasons: Season[];
}

export interface Season {
  number: number;
  episodes: Episode[];
}

export interface Episode {
  season: number;
  number: number;
  title: string;
  plot?: NullableString;
  duration: Date | null;
  castID: number;
  airDate: Date | null;
  languages: string[];
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
