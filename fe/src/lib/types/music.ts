import type { Genre, Media, Keyword } from './media';
import type { UUID, NullableDuration } from './utils';

export interface Album extends Media {
  media_id: UUID;
  name: string;
  album_artists: AlbumArtist[];
  image_paths: string[] | null;
  release_date: Date | string | null;
  genres?: Genre[];
  //  studio?: Studio;
  keywords?: Keyword[];
  duration: NullableDuration | null;
  tracks: Track[];
  //languages?: string[];
}

export type AlbumArtist = {
  artist: UUID
  name: string;
  artist_type: 'individual' | 'group';
}

export interface Track {
  media_id: UUID;
  name: string;
  album_id: UUID;
  duration: number | string;
  lyrics?: string | null;
  //languages?: string[];
  track_number: number;
}
