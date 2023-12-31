import type { NullableInt64, NullableString } from './utils';

export type Member = {
  active: boolean,
  uuid: string,
  memberName: string,
  displayName: NullableString,
  email: string,
  profile_pic: string | null,
  bio: NullableString,
  matrix: NullableString,
  xmpp: NullableString,
  irc: NullableString,
  homepage: NullableString,
  //bookwyrm: "",
  regdate: number | Date,
  roles: MemberRole[],
  // private means federated, but visible only if authenticated
  visibility: 'public' | 'followers_only' | 'local' | 'private',
  followers_uri: string,
  following_uri: string,
  sessionTimeout: NullableInt64,
};

export type MemberRole = 'regular' | 'admin' | 'creator' | 'mod' | 'member' | 'banned';
