import type { UUID } from './utils';

export type Place = {
  id: number;
  kind: string;
  name: string;
  lat: number;
  lon: number;
  country: Country;
};

export type Country = {
  id: number;
  name: string;
  code: string;
};

export type City = {
  uuid: UUID;
  name: string;
  lat: number;
  lon: number;
  country: Country;
};
