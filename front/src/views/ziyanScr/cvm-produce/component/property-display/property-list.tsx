import { defineComponent, computed } from 'vue';
import { cloneDeep, isEmpty } from 'lodash';

import PropertyItem from './property-item';
import './index.scss';

export default defineComponent({
  name: 'PropertyList',
  props: {
    properties: { type: Object, default: null },
    hideEmpty: { type: Boolean, default: false },
    includes: { type: Array, default: () => [] },
    excludes: { type: Array, default: () => [] },
    row: Object,
  },
  emits: ['update:modelValue'],
  setup(props) {
    const visibleProperties = computed(() => {
      const properties = cloneDeep(props.properties);

      Object.keys(properties).forEach((key) => {
        if (props.hideEmpty && isEmpty(properties[key]) && typeof properties[key] !== 'number') {
          delete properties[key];
        }

        if (props.includes.length > 0 && !props.includes.includes(key)) {
          delete properties[key];
        }
        if (props.excludes.length > 0 && props.excludes.includes(key)) {
          delete properties[key];
        }
      });

      return properties;
    });

    return () => (
      <div class='cvm-produce-property-list'>
        {Object.keys(visibleProperties.value).map((key) => {
          return <PropertyItem key={key} k={key} v={visibleProperties.value[key]} row={props.row} />;
        })}
      </div>
    );
  },
});
