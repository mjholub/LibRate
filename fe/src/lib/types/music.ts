import type { Person, Group } from './people';
import type { Genre, Media, Keyword } from './media';
import type { UUID, NullableDuration } from './utils';

export interface Album extends Media {
  media_id: UUID;
  name: string;
  album_artists: AlbumArtists;
  image_paths: string[] | null;
  release_date: Date | string | null;
  genres?: Genre[];
  //  studio?: Studio;
  keywords?: Keyword[];
  duration: NullableDuration | null;
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
  duration: number | string;
  lyrics?: string | null;
  //languages?: string[];
  track_number: number;
}
