import type { Genre, Media } from './media';
import type { City, Place } from './places';
import type { UUID } from './utils';

export interface Creator {
  id: number;
  name: string;
  kind?: string | null;
}

export interface Person extends Creator {
  id: number;
  first_name: string;
  other_names: string[] | null;
  last_name: string;
  nick_names: string[] | null;
  roles: string[] | null;
  works: Media[] | null;
  birth: Date | null; // sql.NullTime in the backend
  death: Date | null;
  website: string | null;
  bio: string | null;
  photos: string[] | null;
  hometown: Place | null;
  residence: Place | null;
  added: Date;
  modified: Date | null;
};

// TODO: add more cast info
export type Cast = {
  actors: Person[];
  directors: Person[];
};

export interface Group extends Creator {
  id: number;
  locations?: Place[] | null;
  name: string;
  active: boolean;
  formed?: Date | null;
  disbanded?: Date | null;
  website?: string | null;
  photos?: string[] | null;
  works?: UUID[] | null;
  members?: Person[] | null;
  primary_genre: Genre | null;
  secondary_genres: Genre[] | null;
  kind?: string | null;
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
