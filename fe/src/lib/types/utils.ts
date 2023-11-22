export type UUID = string;

// for working with Go's sql.NullTime
export type NullableDuration = {
  Time: string;
  Valid: boolean;
}

export type NullableString = {
  String: string;
  Valid: boolean;
}

export type NullableDate = {
  Time: Date;
  Valid: boolean;
}
