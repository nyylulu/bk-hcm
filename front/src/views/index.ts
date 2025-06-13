import business from '@/router/module/business';
import rollingServer from '@/views/rolling-server/route-config';
import greenChannel from '@/views/green-channel/route-config';

export const businessViews = business;

export const platformManagementViews = [...rollingServer, ...greenChannel];
