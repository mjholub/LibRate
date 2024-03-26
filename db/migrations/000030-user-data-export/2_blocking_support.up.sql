CREATE TABLE public.blocks (
  blocker_webfinger varchar NOT NULL REFERENCES public.members(webfinger) ON DELETE CASCADE,
  blockee_webfinger varchar NOT NULL REFERENCES public.members(webfinger) ON DELETE CASCADE,
  created timestamp with time zone NOT NULL DEFAULT now()
);
