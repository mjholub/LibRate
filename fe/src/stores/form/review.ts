import { writable, Writable } from 'svelte/store';
import type { TrackRating, CastRating } from '../../types/review.ts';
import type { Track } from '../../types/music.ts';

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
  submitReview: () => void;
  setFavoriteTrack: (track: Track) => void;
  setTrackRatings: (ratings: TrackRating[]) => void;
  // TODO: implement fetching rating scale from user prefs
  getRatingScale: (userID: number) => number;
}

const  initialState: ReviewStoreState = {
  favoriteTrack: null,
  trackRatings: [],
  castRatings: [],
  reviewText: '',
  wordCount: 0,
  ratingScale: 10, // Default rating scale
};

function createReviewStore() {
  const { subscribe, update } = writable<ReviewStoreState>(initialState);

  return {
  subscribe,
  handleReviewChange: (event: Event) => update((state: ReviewStoreState) => {
    state.reviewText = (event.target as HTMLTextAreaElement).value;

  state.wordCount = state.reviewText.split(/\s+/).length;
  }),

  submitReview: async () => {
  let reviewText, wordCount;

  // Subscribe to the store to get current state
  reviewStore.subscribe(state => {
    reviewText = state.reviewText;
    wordCount = state.wordCount;
  })();

  if (wordCount < 20) {
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
      MediaID: 1, // Replace with the actual media ID
      ReviewText: reviewText
    })
  });

  if (response.ok) {
    alert('Review submitted successfully!');
  } else {
    alert('Failed to submit review.');
  }
  // Reset the store state to initial after submission
  set(initialState);
},

  setFavoriteTrack: (track: Track) => update((state: ReviewStoreState) => {
    state.favoriteTrack = track;
  }),

  setTrackRating: (trackRating: TrackRating) => update((state: ReviewStoreState) => {
    state.trackRatings = state.trackRatings.find((tr) => tr.trackID === trackRating.trackID)
      ? state.trackRatings.filter((tr) => tr.trackID !== trackRating.trackID)
      : [...state.trackRatings, trackRating];
  }),
}
};

export const reviewStore: ReviewStore = createReviewStore(); 
