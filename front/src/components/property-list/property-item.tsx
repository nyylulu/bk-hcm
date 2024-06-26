import { defineComponent, ref, onMounted } from 'vue';
import { useFieldVal } from './field-map';
import { getSubnetConfigs } from '@/api/host/config-management';
export default defineComponent({
  props: {
    k: {
      type: String,
      default: null,
    },
    v: {
      type: [Object, String, Number, Array, Boolean],
      default: null,
    },
    kVisible: {
      type: Boolean,
      default: true,
    },
  },
  setup(props) {
    const { getFieldCn, getFieldCnVal } = useFieldVal();
    const extVal = ref('');
    const getExtVal = (k, v) => {
      if (!['vpc', 'subnet'].includes(k)) return;
      const caseMap = {
        vpc: {
          key: 'vpc_id',
          value: 'vpc_name',
        },
        subnet: {
          key: 'subnet_id',
          value: 'subnet_name',
        },
      };
      const rules = [];
      if (v) {
        rules.push({
          field: caseMap[k].key,
          operator: 'contains',
          value: v,
        });
      }
      const params = {
        page: {
          start: 0,
          limit: 50,
          enable_count: false,
        },
      };
      if (rules.length)
        params.filter = {
          condition: 'AND',
          rules,
        };
      getSubnetConfigs(params).then((res) => {
        extVal.value = res?.data?.info?.[0][caseMap[k].value] || '';
      });
    };
    onMounted(() => {
      getExtVal(props.k, props.v);
    });
    const normal = () => (
      <div class='property-value'>
        {getFieldCnVal(props.k, props.v)}
        {extVal.value ? `| ${extVal.value}` : ''}
      </div>
    );
    return () => (
      <div class='property-item'>
        {props.kVisible ? <div>{getFieldCn(props.k)}</div> : null}
        {normal()}
      </div>
    );
  },
});
