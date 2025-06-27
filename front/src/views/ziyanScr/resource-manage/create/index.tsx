import { PropType, defineComponent, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Button, Checkbox, Dialog, Message, TagInput } from 'bkui-vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import FilterFormItems from '../filter-form-items';
import InputNumber from '@/components/input-number';
import QcloudRegionSelector from '@/views/ziyanScr/components/qcloud-resource/region-selector.vue';
import QcloudZoneSelector from '@/views/ziyanScr/components/qcloud-resource/zone-selector.vue';
import ScrCreateFilterSelector from './ScrCreateFilterSelector';
import CreateRecallTaskDialog from './CreateRecallTaskDialog';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useZiyanScrStore } from '@/store';
import './index.scss';
import type { PaginationType } from '@/hooks/usePagination';
import { IQcloudRegionItem, IQcloudZoneItem } from '@/store/config/qcloud-resource';

interface ILaunchDeviceItem {
  bk_host_id: number;
  asset_id: string;
  ip: string;
  outer_ip: string;
  isp: string;
  device_type: string;
  os_type: string;
  region: string;
  zone: string;
  module: string;
  equipment: number;
  idc_unit: string;
  idc_logic_area: string;
  raid_type: string;
  input_time: string;
}

interface ScrResourceManageCreateFilterType {
  ips: string[];
  asset_ids: string[];
  spec: Spec;
}

interface Spec {
  device_type: string[];
  bk_cloud_regions: string[];
  bk_cloud_zones: string[];
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
        bk_cloud_regions: [],
        bk_cloud_zones: [],
        os_type: [],
      },
    });
    const filter = ref(getDefaultFilter());

    // TODO: 后端上架接口支持bk_cloud_regions, bk_cloud_zones后，这里需要删掉
    const selectedQcloudResource: Record<string, string[]> = { regionNames: [], zoneNames: [] };
    const handleQcloudRegionChange = (qcloudRegions: IQcloudRegionItem[]) => {
      selectedQcloudResource.regionNames = qcloudRegions.map((item) => item.cmdb_region_name);
    };
    const handleQcloudZoneChange = (qcloudZones: IQcloudZoneItem[]) => {
      selectedQcloudResource.zoneNames = qcloudZones.map((item) => item.cmdb_zone_name);
    };

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
          <QcloudRegionSelector
            v-model={filter.value.spec.bk_cloud_regions}
            displayKey={props.type === 'online' ? 'cmdb_region_name' : undefined}
            onChange={handleQcloudRegionChange}
          />
        ),
      },
      {
        label: '园区',
        render: () => (
          <QcloudZoneSelector
            v-model={filter.value.spec.bk_cloud_zones}
            region={filter.value.spec.bk_cloud_regions}
            displayKey={props.type === 'online' ? 'cmdb_zone_name' : undefined}
            onChange={handleQcloudZoneChange}
          />
        ),
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

    const pagination = reactive<PaginationType>({
      start: 0,
      count: 0,
      limit: 10,
      'limit-list': [10, 20, 50, 100, 500],
    });
    const { CommonTable, dataList, getListData } = useTable({
      tableOptions: { columns, extra: { pagination } },
      requestOption: { dataPath: 'data.info', full: true },
      scrConfig: () => {
        // TODO: 上架相关的接口，后端暂时没有支持bk_cloud_regions, bk_cloud_zones，但后面会支持，因此前端这边暂时保留现状，接口处增加判断处理
        const { ips, asset_ids, spec } = filter.value;
        if (props.type === 'online') {
          return {
            url,
            payload: {
              ips,
              asset_ids,
              spec: {
                ...spec,
                region: selectedQcloudResource.regionNames,
                zone: selectedQcloudResource.zoneNames,
                bk_cloud_regions: undefined,
                bk_cloud_zones: undefined,
              },
            },
          };
        }
        return {
          url,
          payload: filter.value,
        };
      },
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
          onlineList.value = commonTableRef.value.tableRef.getSelection();
        }
      } else {
        commonTableRef.value.tableRef.clearSelection();
        onlineList.value = [];
      }
    };

    const isOnlineLoading = ref(false);
    const onlineList = ref<Partial<ILaunchDeviceItem>[]>([]);
    const isOnlineConfirmShow = ref(false);
    const handleOnline = () => {
      onlineList.value = commonTableRef.value.tableRef.getSelection();
      isOnlineConfirmShow.value = true;
    };
    const handleOnlineConfirm = async () => {
      try {
        const bk_host_ids = onlineList.value.map(({ bk_host_id }) => bk_host_id);
        const {
          data: { id },
        } = await ziyanScrStore.createOnlineTask({ bk_host_ids });
        Message({ theme: 'success', message: '提交成功' });
        // 跳转至资源上架详情
        router.push({ name: 'scrResourceManageDetail', params: { id }, query: { type: 'online' } });
      } finally {
        isOnlineLoading.value = false;
      }
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
                  <Button theme='primary' loading={isOnlineLoading.value} onClick={handleOnline}>
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
        <Dialog
          class='online-confirm-dialog'
          isShow={isOnlineConfirmShow.value}
          title='上架确认'
          onConfirm={handleOnlineConfirm}
          onClosed={() => (isOnlineConfirmShow.value = false)}>
          <p>本次上架 {onlineList.value.length} 台，请确认后操作。</p>
          <div class='copy-button-group'>
            <CopyToClipboard content={onlineList.value.map(({ asset_id }) => asset_id).join('\n')}>
              <Button theme='primary'>复制固资号</Button>
            </CopyToClipboard>
            <CopyToClipboard content={onlineList.value.map(({ ip }) => ip).join('\n')}>
              <Button theme='primary'>复制IP</Button>
            </CopyToClipboard>
          </div>
        </Dialog>
        {/* 资源下架 */}
        {props.type === 'offline' && (
          <CreateRecallTaskDialog ref={createRecallTaskDialogRef} onReloadTable={reloadTableDataList} />
        )}
      </div>
    );
  },
});
