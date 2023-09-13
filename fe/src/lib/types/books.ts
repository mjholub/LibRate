import type { Person } from './people';
import type { UUID } from './utils';
import type { Media } from './media';

export interface Book extends Media {
  media_id: UUID;
  title: string;
  authors: Person[];
  publisher: string;
  publication_date: Date;
  genres: string[];
  keywords?: string[];
  languages: string[];
  pages: number;
  isbn?: string;
  asin?: string;
  cover?: string;
  summary: string;
}
