import axios from 'axios'
import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Member } from '$lib/types/member';

export const memberInfo: Member = {
  memberName: '',
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
  checkFollowing: (jwtToken: string, follower: string, followee: string) => Promise<boolean>;
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
    checkFollowing: async (jwtToken: string, follower: string, followee: string) => {
      return new Promise(async (resolve, reject) => {
        const res = await axios.get(`/api/members/is_following/${followee}?follower=${follower}`, {
          headers: {
            'Authorization': `Bearer ${jwtToken}`
          }
        });

        res.status === 200 ? resolve(res.data.message) : reject(Error);
      });
    },
  };
}

export const memberStore: MemberStore = createMemberStore();
