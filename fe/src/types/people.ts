import type { Genre, Media } from './media';
import type { City, Place } from './places';
import type { UUID } from './utils';

export type Person = {
  id: number;
  first_name: string;
  other_names: string[];
  last_name: string;
  nick_names: string[];
  roles: string[];
  works: Media[];
  birth: Date | null; // sql.NullTime in the backend
  death: Date | null;
  website: string;
  bio: string;
  photos: string[];
  hometown: Place;
  residence: Place;
  added: Date;
  modified: Date | null;
};

// TODO: add more cast info
export type Cast = {
  actors: Person[];
  directors: Person[];
};

export interface Group {
  id: number;
  locations?: Place[];
  name: string;
  active: boolean;
  formed?: Date | null;
  disbanded?: Date | null;
  website?: string;
  photos?: string[];
  works?: UUID[];
  members?: Person[];
  primary_genre: Genre;
  secondary_genres: Genre[] | null;
  kind?: string;
  added: Date;
  modified?: Date | null;
  bio: string | null;
  soundcloud: string | null;
  bandcamp: string | null;
  wikipedia: string | null;
}

export interface Studio {
  id: number;
  name: string;
  active: boolean;
  city?: City;
  artists?: Person[];
  works?: Media;
  is_film: boolean;
  is_music: boolean;
  is_tv: boolean;
  is_publishing: boolean;
  is_game: boolean;
}
