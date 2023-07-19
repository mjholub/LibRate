import { writable } from "svelte/store";
import type { Member } from '../../types/member.ts';

export const isAuthenticated = writable(false);
export const member = writable<Member>({
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
  regdate: Date.now(),
  roles: []
});
