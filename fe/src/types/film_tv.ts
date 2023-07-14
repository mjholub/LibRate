import type { Person } from './people';
import type { UUID } from './utils';

export interface Film {
  UUID: UUID;
  kind: 'film';
  title: string;
  created: Date;
  creator: Person;
  // other film specific properties
}

export interface TVShow {
  UUID: UUID;
  kind: 'tvshow';
  title: string;
  created: Date;
  creator: Person;
  // other tvshow specific properties
}
