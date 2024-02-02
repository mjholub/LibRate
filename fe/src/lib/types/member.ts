import type { NullableInt64, NullableString } from './utils';

export type Member = {
  active: boolean,
  webfinger: string,
  uuid: string,
  memberName: string,
  displayName: string,
  email: string,
  profile_pic: string | null,
  bio: string,
  regdate: number | Date,
  roles: MemberRole[],
  visibility: 'public' | 'followers_only' | 'local' | 'private',
  followers_uri: string,
  following_uri: string,
  customFields: Map<string, string>[],
};

export type MemberRole = 'regular' | 'admin' | 'creator' | 'mod' | 'member' | 'banned';
