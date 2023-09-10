import type { PrerenderHttpErrorHandlerValue } from "@sveltejs/kit";

export const prerender = true;
export const trailingSlash = 'ignore';

export const config = {
  params: ['page'],
  handleHttpError: <PrerenderHttpErrorHandlerValue>'warn'
};
