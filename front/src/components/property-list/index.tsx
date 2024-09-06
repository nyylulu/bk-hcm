import { defineComponent, computed } from 'vue';
import PropertyItem from './property-item';
import { cloneDeep, isEmpty } from 'lodash';
import './index.scss';
export default defineComponent({
  name: 'PropertyList',
  components: {
    PropertyItem,
  },
  props: {
    properties: {
      type: Object,
      default: null,
    },
    includes: {
      type: Array,
      default: () => {
        return [];
      },
    },
    hideEmpty: {
      type: Boolean,
      default: false,
    },
    excludes: {
      type: Array,
      default: () => () => {
        return [];
      },
    },
    horizontal: {
      type: Boolean,
      default: false,
    },
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
      <div class='property-list'>
        {Object.keys(visibleProperties.value).map((key) => {
          return <property-item key={key} k={key} v={visibleProperties.value[key]} />;
        })}
      </div>
    );
  },
});
