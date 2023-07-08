export type Media = {
  UUID: string;
  kind: string;
  name: string;
  genres: Genre[];
  keywords: string[];
  lang_ids: number[];
}

export type Genre = {
  id: number;
  media_id: string; // UUID
  name: string;
  desc_short: string;
  desc_long: string;
  keywords: string[]
  parent_genre: Genre;
  children: Genre[];
}
