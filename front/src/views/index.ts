import rollingServer from '@/views/rolling-server/route-config';
import greenChannel from '@/views/green-channel/route-config';

export const platformManagementViews = [...rollingServer, ...greenChannel];
