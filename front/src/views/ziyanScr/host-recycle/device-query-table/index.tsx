import { defineComponent, ref, onMounted } from 'vue';
import { useAccountStore } from '@/store';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getDeviceTypeList, getRegionList, getZoneList, getRecycleStageOpts } from '@/api/host/recycle';
import { Search } from 'bkui-vue/lib/icon';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import dayjs from 'dayjs';

export default defineComponent({
  components: {
    MemberSelect,
    ExportToExcelButton,
  },
  emits: ['goBillDetailPage'],
  setup(props, { emit }) {
    const defaultDeviceForm = () => ({
      bk_biz_id: [],
      device_type: [],
      bk_zone_name: [],
      sub_zone: [],
      stage: [],
      bk_username: [],
    });
    const defaultPartForm = () => ({
      handlerTime: [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()],
      order_id: '',
      suborder_id: '',
      ip: '',
    });
    const deviceForm = ref(defaultDeviceForm());
    const partForm = ref(defaultPartForm());
    const accountStore = useAccountStore();
    const bussinessList = ref([]);
    const getBusinesses = async () => {
      const { data } = await accountStore.getBizListWithAuth();
      bussinessList.value = data || [];
    };
    const deviceTypeList = ref([]);
    const bkZoneNameList = ref([]);
    const subZoneList = ref([]);
    const stageList = ref([]);
    const { columns } = useColumns('deviceQuery');
    const routeBillDetail = (params) => {
      emit('goBillDetailPage', params);
    };
    // 在第三个加子单号，需要跳转到单据详情，未用到路由
    columns.splice(1, 0, {
      label: '子单号',
      field: 'suborder_id',
      width: 80,
      render: ({ row }) => {
        return (
          // 单据详情
          <span class='sub-order-num' onClick={() => routeBillDetail({ pageIndex: 1, params: row.suborder_id })}>
            {row.suborder_id}
          </span>
        );
      },
    });
    const tableColumns = [...columns];
    const pageInfo = ref({
      start: 0,
      limit: 10,
      enable_count: false,
    });
    const requestListParams = ref({
      page: pageInfo.value,
    });
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: tableColumns,
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/task/findmany/recycle/host',
          payload: {
            ...requestListParams.value,
          },
        };
      },
    });
    const getDevicelist = (enableCount = false) => {
      pageInfo.value.enable_count = enableCount;
      const params = {
        ...deviceForm.value,
        start: dayjs(partForm.value.handlerTime[0]).format('YYYY-MM-DD'),
        end: dayjs(partForm.value.handlerTime[1]).format('YYYY-MM-DD'),
        order_id:
          partForm.value.order_id
            .trim()
            .split('|')
            .map((item) => Number(item)) || [],
        suborder_id: partForm.value.suborder_id.trim().split('|'),
        ip: partForm.value.suborder_id.trim().split('|'),
        page: enableCount ? Object.assign(pageInfo.value, { limit: 0 }) : pageInfo.value,
      };
      params.order_id = params.order_id.map((item) => Number(item));

      requestListParams.value = { ...params };
      getListData();
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      getDevicelist(true);
    };
    const clearFilter = () => {
      deviceForm.value = defaultDeviceForm();
      partForm.value = defaultPartForm();
      filterOrders();
    };
    const fetchDeviceTypeList = async () => {
      const data = await getDeviceTypeList();
      deviceTypeList.value = data?.info || [];
    };
    const fetchRegionList = async () => {
      const data = await getRegionList();
      bkZoneNameList.value = data?.info || [];
    };
    const fetchZoneList = async () => {
      const data = await getZoneList();
      subZoneList.value = data?.info || [];
    };
    const fetchStageList = async () => {
      const data = await getRecycleStageOpts();
      stageList.value = data?.info || [];
    };
    onMounted(() => {
      fetchDeviceTypeList();
      fetchRegionList();
      fetchZoneList();
      fetchStageList();
      getBusinesses();
    });
    const deviceRef = ref(null);
    return () => (
      <div>
        <CommonTable ref={deviceRef}>
          {{
            tabselect: () => (
              <bk-form label-width='110' class='bill-filter-form' model={deviceForm}>
                <bk-form-item label='业务'>
                  <bk-select v-model={deviceForm.value.bk_biz_id} multiple clearable placeholder='请选择业务'>
                    {bussinessList.value.map(({ key, value }) => {
                      return <bk-option key={key} label={value} value={key}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='单号'>
                  {/* TODO 旧float-input*/}
                  <bk-input
                    class='order-width'
                    v-model={partForm.value.order_id}
                    clearable
                    placeholder='请输入单号，多个用"|"分割'
                  />
                </bk-form-item>
                <bk-form-item label='子单号'>
                  {/* TODO 旧float-input*/}
                  <bk-input
                    class='order-width'
                    v-model={partForm.value.suborder_id}
                    clearable
                    placeholder='请输入单号，多个用"|"分割'
                  />
                </bk-form-item>
                <bk-form-item label='机型'>
                  <bk-select v-model={deviceForm.value.device_type} multiple clearable placeholder='请选择机型'>
                    {deviceTypeList.value.map((item) => {
                      return <bk-option key={item} label={item} value={item}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='地域'>
                  <bk-select v-model={deviceForm.value.bk_zone_name} multiple clearable placeholder='请选择地域'>
                    {bkZoneNameList.value.map((item) => {
                      return <bk-option key={item} label={item} value={item}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='园区'>
                  <bk-select v-model={deviceForm.value.sub_zone} multiple clearable placeholder='请选择园区'>
                    {subZoneList.value.map((item) => {
                      return <bk-option key={item} label={item} value={item}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='状态'>
                  <bk-select v-model={deviceForm.value.stage} multiple clearable placeholder='请选择状态'>
                    {stageList.value.map(({ stage, description }) => {
                      return <bk-option key={stage} label={description} value={stage}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='回收IP'>
                  {/* TODO 旧float-input*/}
                  <bk-input
                    class='order-width'
                    v-model={partForm.value.ip}
                    clearable
                    placeholder='请输入IP，多个用"|"分割'
                  />
                </bk-form-item>
                <bk-form-item label='回收人'>
                  <member-select
                    class='tag-input-width'
                    v-model={deviceForm.value.bk_username}
                    multiple
                    clearable
                    placeholder='请输入企业微信名'
                  />
                </bk-form-item>
                <bk-form-item label='完成时间'>
                  <bk-date-picker v-model={partForm.value.handlerTime} type='daterange' />
                </bk-form-item>
                <bk-form-item class='bill-form-btn' label-width='20'>
                  <bk-button theme='primary' onClick={filterOrders}>
                    <Search />
                    查询
                  </bk-button>
                  <bk-button onClick={clearFilter}>清空</bk-button>
                  <export-to-excel-button
                    data={deviceRef.value?.dataList}
                    columns={tableColumns}
                    filename='回收设备列表'
                  />
                </bk-form-item>
              </bk-form>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
