import { type PropertyDisplayConfig } from '@/model/typings';

export type AppearanceType = 'status' | 'link' | 'wxwork-link';

export type DisplayType = {
  on?: 'cell' | 'info' | 'search';
  appearance?: AppearanceType;
  showOverflowTooltip?: boolean;
} & PropertyDisplayConfig;
