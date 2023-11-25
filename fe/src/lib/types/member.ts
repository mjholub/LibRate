export type Member = {
  id: number,
  active: boolean,
  uuid: string,
  memberName: string,
  displayName: string | null,
  email: string,
  profilePic: string | null,
  bio: string | null,
  matrix: string | null,
  xmpp: string | null,
  irc: string | null,
  homepage: string | null,
  //bookwyrm: "",
  regdate: number | Date,
  roles: MemberRole[],
  // private means federated, but visible only if authenticated
  visibility: 'public' | 'followers_only' | 'local' | 'private',
};

export type MemberRole = 'regular' | 'admin' | 'creator' | 'mod';
