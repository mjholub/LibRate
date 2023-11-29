import axios from 'axios';
import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Member } from '$lib/types/member.ts';

export const isAuthenticated = writable(false);
export interface AuthStoreState extends Member {
  isAuthenticated: boolean;
};

export type authData = {
  isAuthenticated: boolean;
  memberName: string;
};

interface AuthStore extends Writable<AuthStoreState> {
  authenticate: (token: string) => Promise<authData>;
  logout: () => void;
  getMember: (email_or_username: string) => Promise<Member>;
}

export const initialAuthState: AuthStoreState = {
  id: 0,
  memberName: '',
  displayName: { String: '', Valid: false },
  email: '',
  profilePic: '',
  bio: { String: '', Valid: false },
  matrix: { String: '', Valid: false },
  xmpp: { String: '', Valid: false },
  irc: { String: '', Valid: false },
  homepage: { String: '', Valid: false },
  regdate: new Date(),
  roles: ['member'],
  isAuthenticated: false,
  visibility: 'public',
  followers_uri: '',
  following_uri: '',
  sessionTimeout: { Int64: 0, Valid: false },
  active: false,
  uuid: '',
  publicKeyPem: ''
};


function createAuthStore(): AuthStore {
  const { subscribe, set, update } = writable<AuthStoreState>(initialAuthState);

  return {
    subscribe,
    set,
    update,
    authenticate: async (token: string) => {
      return new Promise<authData>(async (resolve, reject) => {
        try {
          const authStatus = await axios.get(`/api/authenticate/status`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });


          if (authStatus.status === 200) {
            isAuthenticated.set(authStatus.data.isAuthenticated);
            resolve({ isAuthenticated: authStatus.data.isAuthenticated, memberName: authStatus.data.memberName });
          } else {
            // Handle other non-200 status codes here
            isAuthenticated.set(false);
            reject(Error(`Unexpected status code: ${authStatus.status}`));
          }
        } catch (error) {
          // Handle errors from the axios request, including 401 status code
          if (axios.isAxiosError(error) && error.response?.status === 401) {
            isAuthenticated.set(false);
            reject(Error('Unauthorized'));
          } else {
            isAuthenticated.set(false);
            reject(error);
          }
        }
      });
    },
    getMember: async (email_or_username: string) => {
      const res = await fetch(`/api/members/${email_or_username}/info`);
      const resData = await res.json();
      if (resData.message !== "success") {
        throw new Error('Error while retrieving member data');
      }
      const member: Member = resData.data;
      console.debug('member data retrieved from API: ', member);
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

