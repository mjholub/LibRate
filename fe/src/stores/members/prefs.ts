export type MemberPreferences = {
  locale: string | null | undefined;
  auto_accept_follow: boolean;
  locally_searchable: boolean;
  robots_searchable: boolean;
  blur_nsfw: boolean;
  theme: string;
  rating_scale_lower: number;
  rating_scale_upper: number;
  searchable_to_federated: boolean;
  message_autohide_words: string[];
  muted_instances: string[];
};
