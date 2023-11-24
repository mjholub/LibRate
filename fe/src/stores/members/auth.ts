import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Member } from '$lib/types/member.ts';

export const nickName = writable<string>('');
export const isAuthenticated = writable(false);
export interface AuthStoreState extends Member {
  id: number;
  memberName: string;
  roles: string[];
  isAuthenticated: boolean;
};

interface AuthStore extends Writable<AuthStoreState> {
  authenticate: () => Promise<void>;
  logout: () => void;
  getMember: (nickname: string) => Promise<Member>;
}

export const initialAuthState: AuthStoreState = {
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
  isAuthenticated: false,
  visibility: 'public'
};


function createAuthStore(): AuthStore {
  const { subscribe, set, update } = writable<AuthStoreState>(initialAuthState);

  return {
    subscribe,
    set,
    update,
    authenticate: async () => {
      const sessionCookie = document.cookie.includes('session=');
      if (sessionCookie) {
        // using try-cacth to avoid unhandled promise rejection
        try {
          const res = await fetch(`/api/authenticate`);
          res.ok ? isAuthenticated.set(true) : isAuthenticated.set(false);
        } catch (err) {
          isAuthenticated.set(false);
          if (import.meta.env.DEV) {
            console.error('Error while authenticating', err);
          }
        }
      }
      else {
        isAuthenticated.set(false);
        if (import.meta.env.DEV) {
          console.error('Authentication cookie not found');
        }
      }
    },
    getMember: async (nickname: string) => {
      const res = await fetch(`/api/members/${nickname}/info`);
      const member = await res.json();
      member.passhash = '';
      member.id = 0;
      return member;
    },
    logout: () => {
      document.cookie = 'session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
      isAuthenticated.set(false);
    }
  };
}

export const authStore: AuthStore = createAuthStore();

