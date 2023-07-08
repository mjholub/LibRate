import type { Cast } from './people';

export type Review = {
  id: number;
  numstars: number;
  comment: string;
  topic: string;
  attribution: string;
  userid: number;
  mediaid: string; // UUID, use import { v4 as uuid } from 'uuid' in code using this type
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
  rating: Review;
  userid: number;
};

export type ThemeVote = {
  id: number;
  mediaid: number;
  theme: string;
  numstars: number;
  userid: number;
};
