import type { Media } from './media';
import type { Place } from './places';

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
