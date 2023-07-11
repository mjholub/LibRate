import { writable } from 'svelte/store';
import type { UUID } from '../../types/utils.ts';

// create a writable store
export const videoWork = writable(false);

// Function to update videoWork store
export const isVideoWork = async (mediaID: UUID) => {
  const response = await fetch(`/api/media/${mediaID}`);
  const media = await response.json();
  // update the store value
  videoWork.set(media.Kind === 'Film' || media.Kind === 'TV' || media.Kind === 'Anime');
};
