import editItem from '@/components/edit-item';

import { App } from 'vue';

export default {
  install(app: App) {
    app.component(editItem.name, editItem);
  },
};
