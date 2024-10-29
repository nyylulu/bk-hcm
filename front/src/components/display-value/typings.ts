export type AppearanceType = 'status' | 'link';

export type DisplayType = {
  on?: 'cell' | 'info' | 'search';
  appearance?: AppearanceType;
  showOverflowTooltip?: boolean;
};
