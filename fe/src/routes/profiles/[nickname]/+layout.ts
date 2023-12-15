export const ssr = false;
export const prerender = false;
import type { PageServerLoad } from "../$types";

export const load: PageServerLoad = async ({ params }: any) => {
  let { nickname } = params;
  const res = await fetch(`/api/members/${nickname}/info`);
  const profile = await res.json();
  nickname = profile.memberName;
  return { props: { nickname } };
}
