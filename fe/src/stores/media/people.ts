import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Person } from "../../types/people";
import type { UUID } from "../../types/utils";

interface PeopleStoreState {
  people: Person[];
  selectedPerson?: Person;
}

interface PeopleStore extends Writable<PeopleStoreState> {
  getPerson: (id: UUID) => Promise<void>;
  getPeople: () => Promise<void>;
  setPerson: (person: Person) => void;
}

export const initialPerson: Person = {
  id: 0,
  first_name: "",
  other_names: [],
  last_name: "",
  nick_names: [],
  roles: [],
  works: null,
  birth: null,
  death: null,
  website: "",
  bio: "",
  photos: [],
  hometown: null,
  residence: null,
  added: new Date(),
  modified: null
};

const initialState: PeopleStoreState = {
  people: [],
  selectedPerson: initialPerson,
};

function createPeopleStore(): PeopleStore {
  const { subscribe, set, update } = writable<PeopleStoreState>(initialState);

  return {
    subscribe,
    set,
    update,

    getPerson: async (id: UUID) => {
      const response = await fetch(`/api/person/${id}`);
      const person = await response.json();
      update((state: PeopleStoreState) => ({ ...state, selectedPerson: person }));
    },

    getPeople: async () => {
      const response = await fetch(`/api/people`);
      const people = await response.json();
      update((state: PeopleStoreState) => ({ ...state, people }));
    },

    setPerson: (person: Person) => update((state: PeopleStoreState) => {
      state.selectedPerson = person;
      return state;
    }),
  };
}

export const personStore: PeopleStore = createPeopleStore();
