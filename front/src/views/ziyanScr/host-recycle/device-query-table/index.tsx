import { defineComponent, ref, computed, onMounted, watch } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getDeviceTypeList, getRegionList, getZoneList, getRecycleStageOpts } from '@/api/host/recycle';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { Search } from 'bkui-vue/lib/icon';
import BusinessSelector from '@/components/business-selector/index.vue';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import FloatInput from '@/components/float-input';
import dayjs from 'dayjs';
export default defineComponent({
  components: {
    BusinessSelector,
    MemberSelect,
    ExportToExcelButton,
    FloatInput,
  },
  emits: ['goBillDetailPage'],
  setup(props, { emit }) {
    const defaultDeviceForm = () => ({
      bk_biz_id: '',
      order_id: [],
      suborder_id: [],
      ip: [],
      device_type: [],
      bk_zone_name: [],
      sub_zone: [],
      stage: [],
      bk_username: [],
    });
    const defaultTime = () => [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()];
    const deviceForm = ref(defaultDeviceForm());
    watch(
      () => deviceForm.value.bk_biz_id,
      (newVal, oldVal) => {
        if (!oldVal.length) {
          getListData();
        }
      },
    );
    const timeForm = ref(defaultTime());
    const handleTime = (time) => (!time ? '' : dayjs(time).format('YYYY-MM-DD'));
    const timeObj = computed(() => {
      return {
        start: handleTime(timeForm.value[0]),
        end: handleTime(timeForm.value[1]),
      };
    });
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
          <span class='sub-order-num' onClick={() => routeBillDetail(row)}>
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
    const requestListParams = computed(() => {
      const params = {
        ...deviceForm.value,
        ...timeObj.value,
        page: pageInfo.value,
      };
      params.bk_biz_id = params.bk_biz_id === 'all' ? undefined : params.bk_biz_id;
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      removeEmptyFields(params);
      return params;
    });
    const { CommonTable, getListData, dataList } = useTable({
      tableOptions: {
        columns: tableColumns,
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'ip',
          order: 'ASC',
        },
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
    const getDevicelist = () => {
      getListData();
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      getDevicelist();
    };
    const clearFilter = () => {
      const initForm = defaultDeviceForm();
      initForm.bk_biz_id = businessRef.value.defaultBusiness;
      deviceForm.value = initForm;
      timeForm.value = defaultTime();
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
    const businessRef = ref(null);
    onMounted(() => {
      fetchDeviceTypeList();
      fetchRegionList();
      fetchZoneList();
      fetchStageList();
    });
    return () => (
      <div>
        <CommonTable>
          {{
            tabselect: () => (
              <div class={'apply-list-container'}>
                <div class={'filter-container'}>
                  <bk-form label-width='110' class={'scr-form-wrapper'} model={deviceForm}>
                    <bk-form-item label='业务'>
                      <business-selector
                        ref={businessRef}
                        v-model={deviceForm.value.bk_biz_id}
                        placeholder='请选择业务'
                        authed
                        autoSelect
                        clearable={false}
                        isShowAll
                      />
                    </bk-form-item>
                    <bk-form-item label='单号'>
                      <FloatInput v-model={deviceForm.value.order_id} placeholder='请输入单号，多个换行分割' />
                    </bk-form-item>
                    <bk-form-item label='子单号'>
                      <FloatInput v-model={deviceForm.value.suborder_id} placeholder='请输入子单号，多个换行分割' />
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
                      <FloatInput v-model={deviceForm.value.ip} placeholder='请输入IP，多个换行分割' />
                    </bk-form-item>
                    <bk-form-item label='回收人'>
                      <member-select
                        v-model={deviceForm.value.bk_username}
                        multiple
                        clearable
                        placeholder='请输入企业微信名'
                      />
                    </bk-form-item>
                    <bk-form-item label='完成时间'>
                      <bk-date-picker v-model={timeForm.value} type='daterange' />
                    </bk-form-item>
                  </bk-form>
                  <div class='bill-form-btn'>
                    <bk-button theme='primary' onClick={filterOrders}>
                      <Search />
                      查询
                    </bk-button>
                    <bk-button onClick={clearFilter}>清空</bk-button>
                    <export-to-excel-button data={dataList} columns={tableColumns} filename='回收设备列表' />
                  </div>
                </div>
              </div>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
