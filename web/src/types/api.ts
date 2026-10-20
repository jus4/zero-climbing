export type Me = {
  id: number;
  alias: string;
  email: string;
  onboardingRequired: boolean;
};

export type LocationFeatureProperties = {
  title: string;
  description: string;
  user: string;
  user_id: number;
};

export type LocationFeature = {
  type: "Feature";
  id: number;
  properties: LocationFeatureProperties;
  geometry: {
    type: string;
    coordinates: unknown;
  };
};

export type LocationCollection = {
  type: "FeatureCollection";
  features: LocationFeature[];
};

export type ImportCandidate = {
  index: number;
  type: string;
  title: string;
  description: string;
  details: string;
  geojson: unknown;
};

export type ImportUploadResponse = {
  token: string;
  candidates: ImportCandidate[];
};

export type ImportConfirmResponse = {
  imported: number;
};
