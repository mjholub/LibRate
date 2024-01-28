import axios from 'axios'
import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Member } from '$lib/types/member';

export const memberInfo: Member = {
  memberName: '',
  webfinger: '',
  displayName: { String: '', Valid: false },
  email: '',
  profile_pic: '',
  bio: { String: '', Valid: false },
  matrix: { String: '', Valid: false },
  xmpp: { String: '', Valid: false },
  irc: { String: '', Valid: false },
  homepage: { String: '', Valid: false },
  regdate: 0,
  roles: [],
  visibility: "private",
  followers_uri: '',
  following_uri: '',
  sessionTimeout: { Int64: 0, Valid: false },
  active: false,
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

interface MemberStore extends Writable<Member> {
  getMember: (jwtToken: string, email_or_username: string) => Promise<Member>;
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

function createMemberStore(): MemberStore {
  const { subscribe, set, update } = writable<Member>(memberInfo);
  return {
    subscribe,
    set,
    update,
    getMember: async (jwtToken: string, email_or_username: string) => {
      const res = await axios.get(`/api/members/${email_or_username}/info`, {
        headers: {
          'Authorization': `Bearer ${jwtToken}`
        }
      });
      if (res.data.message !== "success") {
        throw new Error('Error while retrieving member data');
      }
      const member: Member = res.data.data;
      console.debug('member data retrieved from API: ', member);
      return member;
    },
    listFollowees: async (jwtToken: string, viewer: string) => {
      return new Promise<Member[]>(async (resolve, reject) => {
        await axios.get(`/api/members/followees`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          },
          params: {
            viewer: viewer
          }
        }).then(res => {
          resolve(res.data.data);
        }).catch(err => {
          reject(err);
        });
      });
    },
    followStatus: async (jwtToken: string | null, followee_webfinger: string) => {
      return new Promise<FollowResponse>(async (resolve, reject) => {
        if (jwtToken === null) {
          resolve({
            id: 0,
            status: 'not_found',
            reblogs: false,
            notify: false,
            acceptTime: null
          }
          );
        }
        await axios.get(`/api/members/follow/status/${followee_webfinger}`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          },
        }).then(res => {
          resolve(res.data.data);
        }).catch(err => {
          reject(err);
        });
      });
    },
    follow: async (req: FollowRequestOut) => {
      return new Promise<FollowResponse>(async (resolve, reject) => {
        await axios.post(`/api/members/follow`, {
          target: req.target,
          reblogs: req.reblogs,
          notify: req.notify
        }, {
          headers: {
            'Authorization': `Bearer ${req.jwtToken}`,
            'X-CSRF-Token': req.CSRFToken
          }
        }).then(res => {
          resolve(res.data.data);
        }).catch(err => {
          reject(err);
        });
      }
      );
    },
    unfollow: async (req: FollowRequestOut) => {
      return new Promise<FollowResponse>(async (resolve, reject) => {
        try {
          const response = await axios.delete(`/api/members/follow`, {
            headers: {
              'Authorization': `Bearer ${req.jwtToken}`,
              'X-CSRF-Token': req.CSRFToken
            },
            data: {
              target: req.target
            }
          });

          resolve(response.data.data);
        } catch (err) {
          reject(err);
        }
      });
    },
    getFollowRequests(jwtToken: string, type: FollowRequestType) {
      return new Promise<FollowRequestIn[] | FollowResponse | FollowRequestsGroup>(async (resolve, reject) => {
        await axios.get(`/api/members/follow/requests/${type}`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          }
        }).then(res => {
          resolve(res.data.data);
        }).catch(err => {
          reject(err);
        });
      });
    },
    cancelFollowRequest: async (jwtToken: string, CSRFToken: string, id: number) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.delete(`/api/members/follow/requests/out/${id}`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`,
            'X-CSRF-Token': CSRFToken
          }
        }).then(_ => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    },
    acceptFollowRequest: async (jwtToken: string, CSRFToken: string, id: number) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.put(`/api/members/follow/requests/in/${id}`, {}, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`,
            'X-CSRF-Token': CSRFToken
          }
        }).then(_ => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    },
    rejectFollowRequest: async (jwtToken: string, CSRFToken: string, id: number) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.delete(`/api/members/follow/requests/in/${id}`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`,
            'X-CSRF-Token': CSRFToken
          }
        }).then(_ => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    },
    block: async (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.post(`/api/members/block`, {
          blocker: blocker_webfinger,
          blockee: blockee_webfinger
        }, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          }
        }).then(_ => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    },
    unblock: async (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.post(`/api/members/unblock`, {
          blocker: blocker_webfinger,
          blockee: blockee_webfinger
        }, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          }
        }).then(_ => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    }
  };
}

export const memberStore: MemberStore = createMemberStore();
