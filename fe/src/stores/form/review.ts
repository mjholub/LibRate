import { writable, get } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { TrackRating, CastRating } from '../../types/review.ts';
import type { Track } from '../../types/music.ts';
import type { UUID } from '../../types/utils.ts';

interface ReviewStoreState {
  favoriteTrack: Track | null;
  trackRatings: TrackRating[] | null;
  castRatings: CastRating[] | null;
  reviewText: string;
  wordCount: number;
  ratingScale: number;
};

interface ReviewStore extends Writable<ReviewStoreState> {
  handleReviewChange: (event: Event) => void;
  submitReview: (mediaID: UUID) => Promise<void>;
  setFavoriteTrack: (track: Track) => void;
  setTrackRatings: (ratings: TrackRating[]) => void;
  setTrackRating: (rating: TrackRating) => void;
  // TODO: implement fetching rating scale from user prefs
  getRatingScale: (userID: number) => number;
}

const initialState: ReviewStoreState = {
  favoriteTrack: null,
  trackRatings: [],
  castRatings: [],
  reviewText: '',
  wordCount: 0,
  ratingScale: 10, // Default rating scale
};

function createReviewStore() {
  const { subscribe, set, update } = writable<ReviewStoreState>(initialState);

  return {
    subscribe,
    handleReviewChange: (event: Event) => update((state: ReviewStoreState) => {
      const reviewText = (event.target as HTMLTextAreaElement).value;
      const wordCount = state.reviewText.split(/\s+/).length;
      return { ...state, reviewText, wordCount };
    }),

    submitReview: async (mediaID: UUID) => {

      // Get the current state of the store synchronously
      const state = get(reviewStore);
      if (!state.wordCount || state.wordCount < 20) {
        alert('Review must be at least 20 words!');
        return;
      }

      const memberID = 1; // Replace with actual member ID
      const response = await fetch('/api/review', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          MemberID: memberID,
          MediaID: mediaID,
          ReviewText: state.reviewText
        })
      });

      response.ok ?
        alert('Review submitted successfully!') :
        alert('Failed to submit review.');
      // Reset the store state to initial after submission
      set(initialState);
    },

    setFavoriteTrack: (track: Track) => update((state: ReviewStoreState) => {
      return { ...state, favoriteTrack: track };
    }),
  
    // setTrackRating handles both adding and removing track ratings
    setTrackRating: (trackRating: TrackRating) => update((state: ReviewStoreState) => {
      // if track is already rated, remove the rating
      // otherwise, add the rating
      if (state.trackRatings) {
        const isRated = state.trackRatings.some((tr) => tr.track.media_id === trackRating.track.media_id);
        const trackRatings = isRated
          ? state.trackRatings.filter((tr) => tr.track.media_id !== trackRating.track.media_id)
          : [...state.trackRatings, trackRating];
        return { ...state, trackRatings };
      } else {
        return { ...state, trackRatings: [trackRating] };
      }
    }),
  }
};

export const reviewStore: ReviewStore = createReviewStore(); 
