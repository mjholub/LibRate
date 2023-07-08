import type { Cast } from './people';
import type { Media } from './media';

export type Review = {
  id: number;
  numstars: number;
  comment: string;
  topic: string;
  attribution: string;
  userid: number;
  mediaid: string; // UUID, use import { v4 as uuid } from 'uuid' in code using this type
  media: Media; // WARN: not fully equivalent to Media type in Go model
  created_at: Date;
  trackratings: TrackRating[];
  castrating: CastRating[];
  themevotes: ThemeVote[];
}

export type TrackRating = {
  id: number;
  track: string;
  rating: number;
};

export type CastRating = {
  id: number;
  mediaid: number;
  cast: Cast;
  numstars: number;
  userid: number;
};

export type ThemeVote = {
  id: number;
  mediaid: number;
  theme: string;
  numstars: number;
  userid: number;
};
