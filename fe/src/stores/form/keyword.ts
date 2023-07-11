import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Keyword } from '../../types/media.ts';
import type { UUID } from '../../types/utils.ts';

interface KeywordStoreState {
  selectedKeywords: Keyword[];
  keywords: Keyword[];
  keywordSearch: string;
}

interface KeywordStore extends Writable<KeywordStoreState> {
  addKeyword: (keyword: Keyword) => void;
  incrementVote: (keyword: Keyword) => void;
  decrementVote: (keyword: Keyword) => void;
  getKeywordsAll: () => Promise<void>;
  getKeywordsByMedia: (mediaID: UUID) => Promise<void>;
  submitKeywordVotes: (mediaID: UUID) => Promise<void>;
  suggestKeywords: (mediaID: UUID, keyword: string) => Promise<void>;
  clearSearch: () => void;
};

const initialState: KeywordStoreState = {
  selectedKeywords: [],
  keywords: [],
  keywordSearch: '',
};

function createKeywordStore() {
  const { subscribe, update } = writable<KeywordStoreState>(initialState);

  return {
    subscribe,
    addKeyword: (keyword: Keyword) => update((state: KeywordStoreState) => {
      state.selectedKeywords = state.selectedKeywords.find((k) => k.keyword === keyword.keyword)
        ? state.selectedKeywords.filter((k) => k.keyword !== keyword.keyword)
        : [...state.selectedKeywords, keyword];
      return state;
    }),

    incrementVote: (keyword: Keyword) => update(state => {
      state.selectedKeywords = state.selectedKeywords.map((k) =>
        k.keyword === keyword.keyword ? { ...k, stars: k.stars + 1 } : k
      );
      return state;
    }),

    decrementVote: (keyword: Keyword) => update(state => {
      state.selectedKeywords = state.selectedKeywords.map((k) =>
        k.keyword === keyword.keyword ? { ...k, stars: k.stars - 1 } : k
      );
      return state;
    }),

    getKeywordsAll: async () => {
      const response = await fetch('/api/keywords/all');
      const keywords = await response.json();
      update((state: KeywordStoreState) => ({ ...state, keywords }));
    },

    getKeywordsByMedia: async (mediaID: UUID) => {
      const response = await fetch(`/api/keywords/media/${mediaID}`);
      const keywords = await response.json();
      update((state: KeywordStoreState) => ({ ...state, keywords }));
    },
submitKeywordVotes: async (mediaID: UUID) => {
  let selectedKeywords;
  
  keywordStore.subscribe((state) => {
    selectedKeywords = state.selectedKeywords;
  })();
  
  const response = await fetch('/api/keywords/vote', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      MediaID: mediaID,
      Keywords: selectedKeywords
    })
  });

  if (response.ok) {
    alert('Keyword votes submitted successfully!');
  } else {
    alert('Failed to submit keyword votes.');
  }
},

suggestKeywords: async (mediaID: UUID, keywordSearch: string) => {
  const response = await fetch('/api/keywords/suggest', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      MediaID: mediaID,
      KeywordSearch: keywordSearch
    })
  });

  if (response.ok) {
    const keywords = await response.json();
    update((state: KeywordStoreState) => ({ ...state, keywords }));
  } else {
    alert('Failed to suggest keywords.');
  }
},
    clearSearch: () => update((state: KeywordStoreState) => ({ ...state, keywordSearch: '' })),
  }
}

export const keywordStore: KeywordStore = createKeywordStore(); 
