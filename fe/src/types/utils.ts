export type UUID = string;

// for working with Go's sql.NullTime
export type NullableDuration = {
  Time: string;
  Valid: boolean;
}
