import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Member } from '../../types/member.ts';

export const memberID = writable(0);
export const isAuthenticated = writable(false);
export interface AuthStoreState extends Member {
  roles: string[];
  isAuthenticated: boolean;
};

interface AuthStore extends Writable<AuthStoreState> {
  authenticate: () => Promise<void>;
  getMember: (memberID: number) => Promise<void>;
}

const initialAuthState: AuthStoreState = {
  id: 0,
  memberName: '',
  displayName: null,
  email: '',
  profilePic: null,
  bio: null,
  matrix: null,
  xmpp: null,
  irc: null,
  homepage: null,
  regdate: new Date(),
  roles: ['member'],
  isAuthenticated: false
};


function createAuthStore(): AuthStore {
  const { subscribe, set, update } = writable<AuthStoreState>(initialAuthState);

  return {
    subscribe,
    set,
    update,
    authenticate: async () => {
      const token = localStorage.getItem('token');
      if (token) {
        const res = await fetch(`/api/authenticate`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        res.ok ? isAuthenticated.set(true) : isAuthenticated.set(false);
      }
    },
    getMember: async (memberID: number) => {
      const res = await fetch(`/api/members/${memberID}`);
      set(await res.json());
    }
  };
}

export const authStore: AuthStore = createAuthStore();

