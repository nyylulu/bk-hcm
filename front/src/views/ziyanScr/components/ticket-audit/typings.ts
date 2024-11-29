import { VNode } from 'vue';

export type ITimelineIconStatusType = 'default' | 'primary' | 'success' | 'warning' | 'danger';

export enum ITimelineNodeType {
  Template = 'template',
  VNode = 'vnode',
}

export interface ITimelineItem {
  tag: string | VNode;
  content?: string | VNode;
  nodeType: ITimelineNodeType;
  type?: ITimelineIconStatusType;
  icon?: VNode;
}
