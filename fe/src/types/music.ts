import type { Person, Studio, Group } from './people';
import type { Genre, Media } from './media';
import type { UUID } from './utils';

export interface Album extends Media {
  media_id: UUID;
  name: string;
  album_artists: AlbumArtists[] | AlbumArtists;
  image_paths: string[] | null;
  release_date: Date;
  genres?: Genre[];
  //  studio?: Studio;
  keywords?: string[];
  duration: number | null;
  tracks: Track[];
  //languages?: string[];
}

type AlbumArtists = {
  person_artist: Person[];
  group_artist: Group[];
}

export interface Track extends Media {
  media_id: UUID;
  name: string;
  album_id: UUID;
  duration: number;
  lyrics?: string | null;
  //languages?: string[];
  track_number: number;
}
