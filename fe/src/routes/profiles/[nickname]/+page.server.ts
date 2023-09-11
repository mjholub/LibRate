import type { PrerenderHttpErrorHandlerValue } from "@sveltejs/kit";
import axios from "axios";

export const load = async (nickname: string) => {
  const { data } = await axios.get(`http://127.0.0.1:3000/api/members/${nickname}/info`);
  console.log(data);
  return {
    props: {
      data
    }
  }
};

export const prerender = false;
export const trailingSlash = 'always';

export const config = {
  params: ['page'],
  handleHttpError: <PrerenderHttpErrorHandlerValue>'warn'
};
