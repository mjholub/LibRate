export type MemberPreferences = {
  ux: UXSettings;
  privsec: PrivacySecurityPreferences;
};

export type UXSettings = {
  locale: string | null | undefined;
  theme: string;
  rating_scale_lower: number;
  rating_scale_upper: number;
}

export type PrivacySecurityPreferences = {
  searchable_to_federated: boolean;
  message_autohide_words: string[];
  muted_instances: string[];
  auto_accept_follow: boolean;
  locally_searchable: boolean;
  robots_searchable: boolean;
  blur_nsfw: boolean;
}