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
          const authStatus = await fetch('/api/authenticate/status?cache=false', {
            method: 'GET',
            headers: {
              'Authorization': `Bearer ${jwtToken}`
            }
          });

          if (authStatus.ok) {
            const data = await authStatus.json();
            isAuthenticated.set(data.isAuthenticated);
            resolve({ isAuthenticated: data.isAuthenticated, memberName: data.memberName });
          } else {
            isAuthenticated.set(false);
            reject(Error(`Unexpected status code: ${authStatus.status}`));
          }
        } catch (error: any) {
          if (typeof error === 'object' && error !== null && 'response' in error) {
            const fetchError = error as { response: { status: number } };
            if (fetchError.response.status === 401) {
              isAuthenticated.set(false);
              reject(Error('Unauthorized'));
            } else {
              isAuthenticated.set(false);
              reject(Error(`Unexpected status code: ${fetchError.response.status}`));
            }
          } else {
            isAuthenticated.set(false);
            reject(error as Error);
          }
        }
      });
    },
    logout: (csrfToken: string) => {
      return new Promise<void>(async (resolve, reject) => {
        const res = await fetch('/api/authenticate/logout', {
          method: 'POST',
          headers: {
            'X-CSRF-Token': csrfToken || ''
          }
        });
        res.ok ? resolve() : reject(Error);
      });
    },
    deleteAccount: async (input: PasswordUpdateInput) => {
      return new Promise<void>(async (resolve, reject) => {
        const res = await fetch('/api/authenticate/delete-account', {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${input.jwtToken}`,
            'X-CSRF-Token': input.csrfToken
          },
          body: JSON.stringify({
            password: input.old,
            confirmation: input.new
          })
        });
        if (res.ok) {
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
        const res = await fetch('/api/authenticate/password', {
          method: 'PATCH',
          headers: {
            Authorization: `Bearer ${input.jwtToken}`,
            'X-CSRF-Token': input.csrfToken
          },
          body: JSON.stringify({
            old: input.old,
            new: input.new,
          })
        });
        res.ok ? resolve() : reject(Error);
      });
    },
  }
}

export const authStore: AuthStore = createAuthStore();

