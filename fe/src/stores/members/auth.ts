import axios from 'axios';
import { writable } from "svelte/store";
import type { Writable } from "svelte/store";

export const isAuthenticated = writable(false);
export interface AuthStoreState {
  isAuthenticated: boolean;
};

export type authData = {
  isAuthenticated: boolean;
  memberName: string;
};

// to avoid writing another type with almost identical fields
// in case of account deletion, 'old' is the password and 'new' is the confirmation
export type PasswordUpdateInput = {
  csrfToken: string;
  jwtToken: string;
  old: string;
  new: string;
}

interface AuthStore extends Writable<AuthStoreState> {
  authenticate: (jwtToken: string) => Promise<authData>;
  logout: (csrfToken: string) => void;
  deleteAccount: (input: PasswordUpdateInput) => void;
  changePassword: (input: PasswordUpdateInput) => Promise<void>;
}

export const initialAuthState: AuthStoreState = {
  isAuthenticated: false,
}


function createAuthStore(): AuthStore {
  const { subscribe, set, update } = writable<AuthStoreState>(initialAuthState);

  return {
    subscribe,
    set,
    update,
    authenticate: async (jwtToken: string) => {
      return new Promise<authData>(async (resolve, reject) => {
        try {

          const authStatus = await axios.get('/api/authenticate/status?cache=false', {
            headers: {
              'Authorization': `Bearer ${jwtToken}`
            }
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
    logout: (csrfToken: string) => {
      return new Promise<void>(async (resolve, reject) => {
        const res = await axios.post(
          '/api/authenticate/logout',
          {},
          {
            headers: {
              'X-CSRF-Token': csrfToken || ''
            }
          }
        );
        res.status === 200 ? resolve() : reject(Error);
      });
    },
    deleteAccount: async (input: PasswordUpdateInput) => {
      return new Promise<void>(async (resolve, reject) => {
        const res = await axios.post('/api/authenticate/delete-account',
          {
            password: input.old,
            confirmation: input.new
          },
          {
            headers: {
              Authorization: `Bearer ${input.jwtToken}`,
              'X-CSRF-Token': input.csrfToken
            }
          });
        if (res.status === 200) {
          authStore.logout(input.csrfToken);
          isAuthenticated.set(false);
          resolve();
        } else {
          reject(Error);
        }
      });
    },
    changePassword: async (input: PasswordUpdateInput) => {
      return new Promise<void>(async (resolve, reject) => {
        const res = await axios.patch('/api/authenticate/password', {
          old: input.old,
          new: input.new,
        },
          {
            headers: {
              Authorization: `Bearer ${input.jwtToken}`,
              'X-CSRF-Token': input.csrfToken
            }
          });
        res.status === 200 ? resolve() : reject(Error);
      });
    },
  }
}

export const authStore: AuthStore = createAuthStore();

