import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import type { Member } from '$lib/types/member';

const memberInfo: Member = {
  id: 0,
  memberName: '',
  displayName: null,
  email: '',
  profilePic: '',
  bio: null,
  matrix: null,
  xmpp: null,
  irc: null,
  homepage: null,
  regdate: 0,
  roles: [],
  visibility: "private",
};

interface MemberStore extends Writable<Member> {
  getMemberByNick: (nick: string) => Promise<Member>;
  //getMemberByID: (id: number) => Promise<Member>;
}

function createMemberStore(): MemberStore {
  const { subscribe, set, update } = writable<Member>(memberInfo);

  return {
    subscribe,
    set,
    update,
    getMemberByNick: async (nick: string) => {
      const res = await fetch(`/api/members/${nick}/info`);
      res.ok || console.error(res.statusText);
      const member = await res.json();
      console.debug('memberStore.getMemberByNick', member);
      return member;
    },
  };
}

export const memberStore: MemberStore = createMemberStore();
