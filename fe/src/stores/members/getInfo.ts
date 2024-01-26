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

interface MemberStore extends Writable<Member> {
  getMember: (jwtToken: string, email_or_username: string) => Promise<Member>;
  follow: (jwtToken: string, follower_webfinger: string, followee_webfinger: string) => Promise<void>;
  unfollow: (jwtToken: string, follower_webfinger: string, followee_webfinger: string) => Promise<void>;
  block: (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => Promise<void>;
  unblock: (jwtToken: string, blocker_webfinger: string, blockee_webfinger: string) => Promise<void>;
  listFollowees: (jwtToken: string, viewer: string) => Promise<Member[]>;
  isFollowing: (jwtToken: string | null, follower_webfinger: string, followee_webfinger: string) => Promise<boolean>;
  verifyViewablity: (jwtToken: string, viewer: string, viewee: string) => Promise<boolean>;
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
    verifyViewablity: async (jwtToken: string, viewer: string, viewee: string) => {
      return new Promise<boolean>(async (resolve, reject) => {
        await axios.get(`/api/members/${viewee}/visibility`, {
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
    isFollowing: async (jwtToken: string | null, follower_webfinger: string, followee_webfinger: string) => {
      return new Promise<boolean>(async (resolve, reject) => {
        if (jwtToken === null) {
          // will return 401 if public follower view is not allowed
          await axios.get(`/api/members/public/is_following`, {
            params: {
              follower: follower_webfinger,
              followee: followee_webfinger
            }
          }).then(res => {
            resolve(res.data.data);
          }).catch(err => {
            reject(err);
          });
        }
        await axios.get(`/api/members/is_following`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          },
          params: {
            follower: follower_webfinger,
            followee: followee_webfinger
          }
        }).then(res => {
          resolve(res.data.data);
        }).catch(err => {
          reject(err);
        });
      });
    },
    follow: async (jwtToken: string, follower_webfinger: string, followee_webfinger: string) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.post(`/api/members/follow`, {
          follower: follower_webfinger,
          followee: followee_webfinger
        }, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          }
        }).then(res => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    },
    unfollow: async (jwtToken: string, follower_webfinger: string, followee_webfinger: string) => {
      return new Promise<void>(async (resolve, reject) => {
        await axios.post(`/api/members/unfollow`, {
          follower: follower_webfinger,
          followee: followee_webfinger
        }, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          }
        }).then(res => {
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
        }).then(res => {
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
        }).then(res => {
          resolve();
        }).catch(err => {
          reject(err);
        });
      });
    }
  };
}

export const memberStore: MemberStore = createMemberStore();
