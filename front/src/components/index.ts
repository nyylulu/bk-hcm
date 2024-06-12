import { App } from 'vue';
import permissionDialog from '@/components/permission-dialog/install-permission';
import editItem from './edit-item/install';

const components = [permissionDialog, editItem];
export default {
  install(app: App) {
    // eslint-disable-next-line array-callback-return
    components.map((item) => {
      app.use(item);
    });
  },
};
