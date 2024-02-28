type resultCategory = "genres" | "members" | "studios" | "ratings" | "artists" | "media" 

export type SearchResponse = {
  categories: resultCategory[];
  totalHits: number;
  processingTime: number;
  page: number;
  totalPages: number;
  hitsPerPage: number;
  data: Map<string, any>;
}
