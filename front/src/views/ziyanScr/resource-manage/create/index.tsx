import { PropType, defineComponent, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Button, Checkbox, Message, TagInput } from 'bkui-vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import FilterFormItems from '../filter-form-items';
import InputNumber from '@/components/input-number';
import ScrIdcZoneSelector from './ScrIdcZoneSelector';
import ScrCreateFilterSelector from './ScrCreateFilterSelector';
import CreateRecallTaskDialog from './CreateRecallTaskDialog';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
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
    const router = useRouter();
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
          <ScrCreateFilterSelector v-model={filter.value.spec.device_type} api={ziyanScrStore.getDeviceTypeList} />
        ),
      },
      {
        label: '操作系统',
        render: () => (
          <ScrCreateFilterSelector v-model={filter.value.spec.os_type} api={ziyanScrStore.getIdcpmOsTypeList} />
        ),
        hidden: props.type === 'offline',
      },
      {
        label: '地域',
        render: () => (
          <ScrCreateFilterSelector v-model={filter.value.spec.region} api={ziyanScrStore.getIdcRegionList} />
        ),
      },
      {
        label: '园区',
        render: () => <ScrIdcZoneSelector v-model={filter.value.spec.zone} cmdbRegionName={filter.value.spec.region} />,
      },
      {
        label: '内网 IP',
        render: () => (
          <TagInput
            v-model={filter.value.ips}
            allow-create
            collapse-tags
            pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
          />
        ),
        hidden: props.type === 'offline',
      },
      {
        label: '固资号',
        render: () => (
          <TagInput
            v-model={filter.value.asset_ids}
            allow-create
            collapse-tags
            pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
          />
        ),
        hidden: props.type === 'offline',
      },
    ];

    const columnName = props.type === 'online' ? 'scrResourceOnlineCreate' : 'scrResourceOfflineCreate';
    const url =
      props.type === 'online'
        ? '/api/v1/woa/pool/findmany/launch/match/device'
        : '/api/v1/woa/pool/findmany/recall/match/device';
    const { columns } = useColumns(columnName);

    props.type === 'offline' &&
      columns.push({
        label: '操作',
        render: ({ data }: any) => (
          <Button
            text
            theme='primary'
            onClick={() => {
              const { amount, ...rest } = data;
              createRecallTaskDialogRef.value.triggerShow(true, { ...rest, replicas: amount });
            }}>
            发起下架
          </Button>
        ),
      });

    const isAutoCheck = ref(false);

    const { CommonTable, dataList, getListData, pagination } = useTable({
      tableOptions: { columns },
      requestOption: { dataPath: 'data.info', full: true },
      scrConfig: () => ({
        url,
        payload: filter.value,
      }),
    });

    const reloadTableDataList = () => {
      pagination.start = 0;
      getListData();
      isAutoCheck.value = false;
    };

    const clearFilter = () => {
      filter.value = getDefaultFilter();
      reloadTableDataList();
    };

    const createRecallTaskDialogRef = ref();
    const commonTableRef = ref();
    const selectNumber = ref(0);

    const handleCheck = (value: boolean) => {
      if (value) {
        for (let i = 0; i < selectNumber.value; i++) {
          commonTableRef.value.tableRef.toggleRowSelection(dataList.value[i]);
        }
      } else {
        commonTableRef.value.tableRef.clearSelection();
      }
    };
    const handleOnline = async () => {
      const ids = commonTableRef.value.tableRef.getSelection().map((item: any) => item.bk_host_id);
      const {
        data: { id },
      } = await ziyanScrStore.createOnlineTask({ bk_host_ids: ids });
      Message({ theme: 'success', message: '提交成功' });
      // 跳转至资源上架详情
      router.push({ name: 'scrResourceManageDetail', params: { id }, query: { type: 'online' } });
    };

    return () => (
      <div class='scr-resource-manage-create-page'>
        <DetailHeader>
          <span class='header-title-prefix'>新增{props.type === 'online' ? '上架' : '下架'}</span>
        </DetailHeader>
        <div class='common-sub-main-container'>
          <div class='sub-main-content'>
            <div class='operation-bar'>
              <FilterFormItems config={filterFormItems} handleSearch={reloadTableDataList} handleClear={clearFilter}>
                {/* 资源下架 */}
                {(function () {
                  if (props.type === 'offline') {
                    return {
                      end: () => (
                        <div class='filter-item'>
                          <Button theme='primary' onClick={() => createRecallTaskDialogRef.value.triggerShow(true)}>
                            发起下架
                          </Button>
                        </div>
                      ),
                    };
                  }
                })()}
              </FilterFormItems>
              {/* 资源上架 */}
              {props.type === 'online' && (
                <div class='table-operation-bar'>
                  <span class='mr8'>数量</span>
                  <InputNumber v-model={selectNumber.value} class='mr8' min={0} max={500} />
                  <Checkbox v-model={isAutoCheck.value} class='mr8' onChange={handleCheck}>
                    一键勾选
                  </Checkbox>
                  <Button theme='primary' onClick={handleOnline}>
                    提交上架
                  </Button>
                </div>
              )}
            </div>
            <CommonTable
              ref={commonTableRef}
              style={{ height: props.type === 'online' ? 'calc(100% - 250px)' : 'calc(100% - 125px)' }}
            />
          </div>
        </div>
        {/* 资源下架 */}
        {props.type === 'offline' && (
          <CreateRecallTaskDialog ref={createRecallTaskDialogRef} onReloadTable={reloadTableDataList} />
        )}
      </div>
    );
  },
});
