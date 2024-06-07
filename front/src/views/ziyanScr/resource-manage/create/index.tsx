import { PropType, defineComponent, ref } from 'vue';
import { Button, Checkbox, TagInput } from 'bkui-vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import FilterFormItems from '../filter-form-items';
import InputNumber from '@/components/input-number';
import ScrIdcZoneSelector from './ScrIdcZoneSelector';
import ScrCreateFilterSelector from './ScrCreateFilterSelector';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useZiyanScrStore } from '@/store';
import './index.scss';

interface ScrResourceManageCreateFilterType {
  ips: string[];
  asset_ids: string[];
  spec: Spec;
}

interface Spec {
  device_type: string[];
  region: string[];
  zone: string[];
  os_type: string[];
}

export default defineComponent({
  name: 'ScrResourceManageCreate',
  props: { type: String as PropType<'online' | 'offline'> },
  setup(props) {
    const ziyanScrStore = useZiyanScrStore();

    const getDefaultFilter = (): ScrResourceManageCreateFilterType => ({
      ips: [],
      asset_ids: [],
      spec: {
        device_type: [],
        region: [],
        zone: [],
        os_type: [],
      },
    });
    const filter = ref(getDefaultFilter());
    const filterFormItems = [
      {
        label: '机型',
        render: () => (
          <ScrCreateFilterSelector
            v-model={filter.value.spec.device_type}
            api={ziyanScrStore.getDeviceTypeList}
            class='w200'
          />
        ),
      },
      {
        label: '操作系统',
        render: () => (
          <ScrCreateFilterSelector
            v-model={filter.value.spec.os_type}
            api={ziyanScrStore.getIdcpmOsTypeList}
            class='w200'
          />
        ),
        hidden: props.type === 'offline',
      },
      {
        label: '地域',
        render: () => (
          <ScrCreateFilterSelector
            v-model={filter.value.spec.region}
            api={ziyanScrStore.getIdcRegionList}
            class='w200'
          />
        ),
      },
      {
        label: '园区',
        render: () => (
          <ScrIdcZoneSelector v-model={filter.value.spec.zone} cmdbRegionName={filter.value.spec.region} class='w200' />
        ),
      },
      {
        label: '内网 IP',
        render: () => <TagInput v-model={filter.value.ips} class='w200' allow-create collapse-tags />,
        hidden: props.type === 'offline',
      },
      {
        label: '固资号',
        render: () => <TagInput v-model={filter.value.asset_ids} class='w200' allow-create collapse-tags />,
        hidden: props.type === 'offline',
      },
    ];

    const columnName = props.type === 'online' ? 'scrResourceOnlineCreate' : 'scrResourceOfflineCreate';
    const url =
      props.type === 'online'
        ? '/api/v1/woa/pool/findmany/launch/match/device'
        : '/api/v1/woa/pool/findmany/recall/match/device';
    const { columns } = useColumns(columnName);

    const { CommonTable, getListData, pagination } = useTable({
      tableOptions: { columns },
      requestOption: { dataPath: 'data.info' },
      scrConfig: () => ({
        url,
        payload: filter.value,
      }),
    });

    const reloadTableDataList = () => {
      pagination.start = 0;
      getListData();
    };

    const clearFilter = () => {
      filter.value = getDefaultFilter();
      reloadTableDataList();
    };

    return () => (
      <div class='scr-resource-manage-create-page'>
        <DetailHeader>
          <span class='header-title-prefix'>新增{props.type === 'online' ? '上架' : '下架'}</span>
        </DetailHeader>
        <div class='common-sub-main-container'>
          <div class='sub-main-content'>
            <div class='operation-bar'>
              <FilterFormItems config={filterFormItems} handleSearch={reloadTableDataList} handleClear={clearFilter} />
              {props.type === 'online' && (
                <div class='table-operation-bar'>
                  <span class='mr8'>数量</span>
                  <InputNumber class='mr8' />
                  <Checkbox class='mr8'>一键勾选</Checkbox>
                  <Button theme='primary'>提交上架</Button>
                </div>
              )}
              <CommonTable />
            </div>
          </div>
        </div>
      </div>
    );
  },
});
