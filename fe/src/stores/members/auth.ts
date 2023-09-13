import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Member } from '$lib/types/member.ts';

export const memberID = writable<number>(0);
export const isAuthenticated = writable(false);
export interface AuthStoreState extends Member {
  id: number;
  roles: string[];
  isAuthenticated: boolean;
};

interface AuthStore extends Writable<AuthStoreState> {
  authenticate: () => Promise<void>;
  getMember: (memberID: number) => Promise<Member>;
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
      if (typeof localStorage !== 'undefined') {
        const token = localStorage.getItem('token');
        if (token) {
          const res = await fetch(`/api/authenticate`, {
            headers: { 'Authorization': `Bearer ${token}` }
          })
          res.ok ? isAuthenticated.set(true) : isAuthenticated.set(false);
        } else {
          isAuthenticated.set(false);
          throw new Error('No token found');
        }
      }
    },
    getMember: async (memberID: number) => {
      const res = await fetch(`/api/members/${memberID}`);
      const member = await res.json();
      return member;
    }
  };
}

export const authStore: AuthStore = createAuthStore();

