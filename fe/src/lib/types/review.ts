import type { Cast } from './people';
import type { Media } from './media';
import type { Track } from './music';

export type Review = {
  id: number;
  numstars: number;
  comment: string;
  topic: string;
  attribution: string;
  userid: number;
  mediaid: string; // UUID, use import { v4 as uuid } from 'uuid' in code using this type
  media: Media;
  created_at: Date;
  trackratings: TrackRating[];
  castrating: CastRating[];
}

export type TrackRating = {
  id: number;
  track: Track;
  rating: number;
};

export type CastRating = {
  id: number;
  mediaid: number;
  cast: Cast;
  numstars: number;
  userid: number;
};
