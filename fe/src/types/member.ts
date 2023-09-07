export type Member = {
  id: number,
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
  roles: string[],
};
