import type { Person } from './people';
import type { UUID } from './utils';

export interface Book {
  media_id: UUID;
  title: string;
  authors: Person[] | null;
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
