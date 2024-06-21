import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Member } from '$lib/types/member';

export const memberInfo: Member = {
  memberName: '',
  webfinger: '',
  displayName: '',
  email: '',
  profile_pic: '',
  bio: '',
  regdate: 0,
  roles: [],
  visibility: "private",
  followers_uri: '',
  following_uri: '',
  active: false,
  customFields: [],
  uuid: ''
};

export type FollowRequestType = 'sent' | 'received' | 'all';

export type FollowRequestOut = {
  jwtToken: string;
  target: string;
  reblogs: boolean;
  notify: boolean;
  CSRFToken: string;
}

export type FollowResponse = {
  status: 'accepted' | 'pending' | 'failed' | 'not_found' | 'already_following' | 'blocked';
  id: number;
  reblogs: boolean;
  notify: boolean;
  acceptTime: Date | null;
}

export interface FollowRequestIn {
  id: number;
  requester: string;
  created: Date;
};

export type FollowRequestsGroup = {
  sent: FollowRequestIn[];
  received: FollowRequestIn[];
}

// TODO: consider changing Writables<T> to another generic type than member
interface FollowStore extends Writable<Member> {
  follow: (req: FollowRequestOut) => Promise<FollowResponse>;
  //updateFollow: (req: FollowRequestIn) => Promise<void>;
  unfollow: (req: FollowRequestOut) => Promise<FollowResponse>;
  cancelFollowRequest: (jwtToken: string, CSRFToken: string, id: number) => Promise<void>;
  acceptFollowRequest: (jwtToken: string, CSRFToken: string, id: number) => Promise<void>;
  rejectFollowRequest: (jwtToken: string, CSRFToken: string, id: number) => Promise<void>;
  getFollowRequests: (jwtToken: string, type: FollowRequestType) => Promise<FollowRequestIn[] | FollowResponse | FollowRequestsGroup>;
  block: (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => Promise<void>;
  unblock: (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => Promise<void>;
  listFollowees: (jwtToken: string, viewer: string) => Promise<Member[]>;
  followStatus: (jwtToken: string | null, followee_webfinger: string) => Promise<FollowResponse>;
}

function createFollowStore(): FollowStore {
  const { subscribe, set, update } = writable<Member>({} as Member);
  return {
    subscribe,
    set,
    update,
    listFollowees: async (jwtToken: string, viewer: string) => {
      const response = await fetch(`/api/members/followees?viewer=${viewer}`, {
        headers: {
          'Authorization': `Bearer ${jwtToken}`
        }
      });
      const data = await response.json();
      return data.data;
    },
    followStatus: async (jwtToken: string | null, followee_webfinger: string) => {
      if (jwtToken === null) {
        return {
          id: 0,
          status: 'not_found',
          reblogs: false,
          notify: false,
          acceptTime: null
        };
      }
      const response = await fetch(`/api/members/follow/status/${followee_webfinger}`, {
        headers: {
          'Authorization': `Bearer ${jwtToken}`
        }
      });
      const data = await response.json();
      return data.data;
    },
    follow: async (req: FollowRequestOut) => {
      const response = await fetch(`/api/members/follow`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${req.jwtToken}`,
          'X-CSRF-Token': req.CSRFToken,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          target: req.target,
          reblogs: req.reblogs,
          notify: req.notify
        })
      });
      const data = await response.json();
      return data.data;
    },
    unfollow: async (req: FollowRequestOut) => {
      const response = await fetch(`/api/members/follow`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${req.jwtToken}`,
          'X-CSRF-Token': req.CSRFToken,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          target: req.target
        })
      });
      const data = await response.json();
      return data.data;
    },
    getFollowRequests: async (jwtToken: string, type: FollowRequestType) => {
      const response = await fetch(`/api/members/follow/requests/${type}`, {
        headers: {
          'Authorization': `Bearer ${jwtToken}`
        }
      });
      const data = await response.json();
      return data.data;
    },
    cancelFollowRequest: async (jwtToken: string, CSRFToken: string, id: number) => {
      await fetch(`/api/members/follow/requests/out/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${jwtToken}`,
          'X-CSRF-Token': CSRFToken
        }
      });
    },
    acceptFollowRequest: async (jwtToken: string, CSRFToken: string, id: number) => {
      await fetch(`/api/members/follow/requests/in/${id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${jwtToken}`,
          'X-CSRF-Token': CSRFToken
        }
      });
    },
    rejectFollowRequest: async (jwtToken: string, CSRFToken: string, id: number) => {
      return fetch(`/api/members/follow/requests/in/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${jwtToken}`,
          'X-CSRF-Token': CSRFToken
        }
      }).then(response => {
        if (!response.ok) {
          return Promise.reject(new Error('Failed to reject follow request'));
        }
      });
    },
    block: async (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => {
      return fetch(`/api/members/block`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${jwtToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          blocker: blocker_webfinger,
          blockee: blockee_webfinger
        })
      }).then(response => {
        if (!response.ok) {
          return Promise.reject(new Error('Failed to block member'));
        }
      });
    },
    unblock: async (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => {
      return fetch(`/api/members/unblock`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${jwtToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          blocker: blocker_webfinger,
          blockee: blockee_webfinger
        })
      }).then(response => {
        if (!response.ok) {
          return Promise.reject(new Error('Failed to unblock member'));
        }
      });
    }
  };
}


export const followStore: FollowStore = createFollowStore();
