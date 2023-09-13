import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { UUID } from '$lib/types/utils.ts';
import type { Album, Track } from '$lib/types/music.ts';

interface AlbumStoreState {
  album: Album | null;
};

interface TrackStoreState {
  track: Track | null;
};

interface AlbumStore extends Writable<AlbumStoreState> {
  getAlbum: (albumID: UUID) => Promise<void>;
};

interface TrackStore extends Writable<TrackStoreState> {
  getTrack: (trackID: UUID) => Promise<void>;
};

const initialAlbumState: AlbumStoreState = {
  album: null,
};

const initialTrackState: TrackStoreState = {
  track: null,
};

function createAlbumStore(): AlbumStore {
  const { subscribe, set, update } = writable<AlbumStoreState>(initialAlbumState);

  return {
    subscribe,
    set,
    update,
    getAlbum: async (albumID: UUID) => {
      const response = await fetch(`/api/album/${albumID}`);
      const album: Album = await response.json();
      set({ ...initialAlbumState, album });
    },
  };
};

function createTrackStore(): TrackStore {
  const { subscribe, set, update } = writable<TrackStoreState>(initialTrackState);

  return {
    subscribe,
    set,
    update,
    getTrack: async (trackID: UUID) => {
      const response = await fetch(`/api/track/${trackID}`);
      const track: Track = await response.json();
      set({ ...initialTrackState, track });
    },
  };
};

export const albumStore: AlbumStore = createAlbumStore();
export const trackStore: TrackStore = createTrackStore();
