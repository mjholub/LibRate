import type { Media } from './media';

type resultCategory = "genres" | "members" | "studios" | "ratings" | "artists" | "media" 

export type SearchResult = {
  id: number;
  category: resultCategory;
  data: Map<string, any>;
};
