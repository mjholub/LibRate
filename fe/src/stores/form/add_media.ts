/*
 * Using the form/ folder, because media/ already has quite a few files.
 * This file is for the "Add Media" form.
 * In the production setting it must be kept behind a protected route.
 */

import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { UUID } from '../../types/utils';
import type { Media } from '../../types/media.ts';

export interface submissionFormState extends Media {
  UUID: UUID,
  kind: 'book' | 'film' | 'game' | 'album' | 'show' | 'anime' | 'manga' | 'comic',
};

const initialState: submissionFormState = {
  UUID: '',
  kind: 'album',
  title: '',
  created: new Date(),
  creator: null,
  creators: [],
  added: new Date(),
  modified: undefined,
};


interface SubmitMediaForm extends Writable<Media> {
  handleMediaChange: (event: Event) => void;
  submitMedia: (media: Media) => Promise<void>;
  updateMedia: (media: Media) => void;
  // searchMedia: (query: string) => Promise<Media[]>;
  // selectArtist: (artist: Media) => void;
  // findArtist: (artist: string) => Promise<Person[] | Group[]>;
};

function createSubmitMediaForm(): SubmitMediaForm {
  const { subscribe, set, update } = writable<Media>(initialState);

  return {
    subscribe,
    set,
    update,
    handleMediaChange: (event: Event) => update((state: Media) => {
      const { name, value } = event.target as HTMLInputElement;
      return { ...state, [name]: value };
    }),
    // submitMedia makes requests to /api/form/add_media/:type (POST), where type is read from the
    // media.kind field of the JSON form data. The backend code then uses the type to determine
    // which table to insert into.
    submitMedia: async (media: Media) => {
      const response = await fetch(`/api/form/add_media/${media.kind}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(media)
      });
      if (response.ok) {
        alert('Media submitted!');
      } else {
        alert('Media submission failed!');
      }
    },
    // updateMedia is not for handling state changes, but making requests to the backend to update
    // the media entry. It is used in the "Edit Media" form.
    updateMedia: async (media: Media) => {
      const response = await fetch(`/api/form/update_media/${media.kind}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(media)
      });
      if (response.ok) {
        alert('Media updated!');
      } else {
        alert('Media update failed!');
      }
    },
  };
}

export const submitMediaForm = createSubmitMediaForm();
