import { defineComponent, ref, onMounted } from 'vue';
import { useFieldVal } from '@/views/ziyanScr/cvm-produce/component/property-display/field-map';
import { QueryFilterTypeLegacy, QueryRuleOPEnumLegacy } from '@/typings';
import rollRequest from '@blueking/roll-request';
import http from '@/http';

interface ICvmSubnetItem {
  id: number;
  region: string;
  zone: string;
  vpc_id: string;
  vpc_name: string;
  subnet_id: string;
  subnet_name: string;
  enable: boolean;
  comment: string;
  [key: string]: any;
}

export default defineComponent({
  props: {
    k: { type: String },
    v: { type: [Object, String, Number, Array, Boolean] },
    kVisible: { type: Boolean, default: true },
    row: Object,
  },
  setup(props) {
    const { getFieldCn, getFieldCnVal } = useFieldVal();

    const reqKeyList = ['vpc', 'subnet'];
    const keyMap: Record<string, { reqKey: string; nameKey: string }> = {
      vpc: { reqKey: 'vpc_id', nameKey: 'vpc_name' },
      subnet: { reqKey: 'subnet_id', nameKey: 'subnet_name' },
    };

    const displayName = ref('');
    const getDisplayName = async (k: string, v: string) => {
      if (!v) return;
      const { spec } = props.row;
      const { region, zone } = spec;

      const filter: QueryFilterTypeLegacy = {
        condition: 'AND',
        rules: [
          { field: 'region', operator: QueryRuleOPEnumLegacy.EQ, value: region },
          { field: 'zone', operator: QueryRuleOPEnumLegacy.EQ, value: zone },
          { field: keyMap[k].reqKey, operator: QueryRuleOPEnumLegacy.EQ, value: v },
        ],
      };

      const list = (await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'enable_count',
      }).rollReqUseCount(
        '/api/v1/woa/config/findmany/config/cvm/subnet/list',
        { filter },
        { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.info },
      )) as ICvmSubnetItem[];

      displayName.value = list.find((item) => item[keyMap[k].reqKey] === v)?.[keyMap[k].nameKey] || '';
    };

    onMounted(() => {
      if (reqKeyList.includes(props.k)) {
        getDisplayName(props.k, props.v as string);
      }
    });

    return () => (
      <div class='cvm-produce-property-item'>
        {props.kVisible ? <div>{getFieldCn(props.k)}：</div> : null}
        <div class='cvm-produce-property-value'>
          {getFieldCnVal(props.k, props.v, props.row)}
          {displayName.value ? `(${displayName.value})` : ''}
        </div>
      </div>
    );
  },
});
