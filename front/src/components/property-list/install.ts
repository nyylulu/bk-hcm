import PropertyList from '@/components/property-list';

import { App } from 'vue';

export default {
  install(app: App) {
    app.component(PropertyList.name, PropertyList);
  },
};
