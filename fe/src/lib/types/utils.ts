export type UUID = string;

// for working with Go's sql.NullTime
// TODO: install sql types
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

export type NullableInt64 = {
  Int64: number;
  Valid: boolean;
}
