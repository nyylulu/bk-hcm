import business from '@/router/module/business';
import task from '@/views/task/route-config';
import rollingServer from '@/views/rolling-server/route-config';
import greenChannel from '@/views/green-channel/route-config';

business.forEach((group) => {
  const index = group.children.findIndex((menu) => menu.name === 'businessRecord');
  if (index !== -1) {
    group.children.splice(index + 1, 0, ...task);
  }
});
export const businessViews = business;

export const platformManagementViews = [...rollingServer, ...greenChannel];
