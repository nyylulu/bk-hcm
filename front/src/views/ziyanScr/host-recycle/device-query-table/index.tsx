import { defineComponent, ref, computed, onMounted, watch } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getDeviceTypeList, getRegionList, getZoneList, getRecycleStageOpts } from '@/api/host/recycle';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { Search } from 'bkui-vue/lib/icon';
import { useUserStore } from '@/store';
import BusinessSelector from '@/components/business-selector/index.vue';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import FloatInput from '@/components/float-input';
import dayjs from 'dayjs';
import { Button, DatePicker, Form, Select } from 'bkui-vue';
const { FormItem } = Form;
export default defineComponent({
  components: {
    BusinessSelector,
    MemberSelect,
    ExportToExcelButton,
    FloatInput,
  },
  emits: ['goBillDetailPage'],
  setup(_, { emit }) {
    const userStore = useUserStore();
    const defaultDeviceForm = () => ({
      bk_biz_id: [],
      order_id: [],
      suborder_id: [],
      ip: [],
      device_type: [],
      bk_zone_name: [],
      sub_zone: [],
      stage: [],
      bk_username: [userStore.username],
    });
    const defaultTime = () => [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()];
    const deviceForm = ref(defaultDeviceForm());
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
        bk_biz_id:
          deviceForm.value.bk_biz_id.length === 0
            ? businessRef.value.businessList.slice(1).map((item: any) => item.id)
            : deviceForm.value.bk_biz_id,
      };
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      removeEmptyFields(params);
      return params;
    });
    const { CommonTable, getListData, dataList, pagination } = useTable({
      tableOptions: {
        columns: tableColumns,
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'ip',
          order: 'ASC',
        },
        immediate: false,
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
    const filterOrders = () => {
      pagination.start = 0;
      deviceForm.value.bk_biz_id =
        deviceForm.value.bk_biz_id.length === 1 && deviceForm.value.bk_biz_id[0] === 'all'
          ? []
          : deviceForm.value.bk_biz_id;
      getListData();
    };
    const clearFilter = () => {
      const initForm = defaultDeviceForm();
      // 因为要保存业务全选的情况, 所以这里 defaultBusiness 可能是 ['all'], 而组件的全选对应着 [], 所以需要额外处理
      // 根源是此处的接口要求全选时携带传递所有业务id, 所以需要与空数组做区分
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

    watch(
      () => userStore.username,
      (username) => {
        deviceForm.value.bk_username = [username];
      },
    );

    watch(
      () => businessRef.value?.businessList,
      (val) => {
        if (!val?.length) return;
        getListData();
      },
      { deep: true },
    );

    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'}>
          <Form formType='vertical' class={'scr-form-wrapper'} model={deviceForm}>
            <FormItem label='业务'>
              <business-selector
                ref={businessRef}
                v-model={deviceForm.value.bk_biz_id}
                placeholder='请选择业务'
                authed
                autoSelect
                clearable={false}
                isShowAll
                notAutoSelectAll
                multiple
                url-key='scr_host_bizs'
                base64Encode
              />
            </FormItem>
            <FormItem label='单号'>
              <FloatInput v-model={deviceForm.value.order_id} placeholder='请输入单号，多个换行分割' />
            </FormItem>
            <FormItem label='子单号'>
              <FloatInput v-model={deviceForm.value.suborder_id} placeholder='请输入子单号，多个换行分割' />
            </FormItem>
            <FormItem label='机型'>
              <Select v-model={deviceForm.value.device_type} multiple clearable placeholder='请选择机型'>
                {deviceTypeList.value.map((item) => {
                  return <Select.Option key={item} name={item} id={item} />;
                })}
              </Select>
            </FormItem>
            <FormItem label='地域'>
              <Select v-model={deviceForm.value.bk_zone_name} multiple clearable placeholder='请选择地域'>
                {bkZoneNameList.value.map((item) => {
                  return <Select.Option key={item} name={item} id={item} />;
                })}
              </Select>
            </FormItem>
            <FormItem label='园区'>
              <Select v-model={deviceForm.value.sub_zone} multiple clearable placeholder='请选择园区'>
                {subZoneList.value.map((item) => {
                  return <Select.Option key={item} name={item} id={item} />;
                })}
              </Select>
            </FormItem>
            <FormItem label='状态'>
              <Select v-model={deviceForm.value.stage} multiple clearable placeholder='请选择状态'>
                {stageList.value.map(({ stage, description }) => {
                  return <Select.Option key={stage} name={description} id={stage} />;
                })}
              </Select>
            </FormItem>
            <FormItem label='回收IP'>
              <FloatInput v-model={deviceForm.value.ip} placeholder='请输入IP，多个换行分割' />
            </FormItem>
            <FormItem label='回收人'>
              <member-select
                v-model={deviceForm.value.bk_username}
                multiple
                clearable
                defaultUserlist={[
                  {
                    username: userStore.username,
                    display_name: userStore.username,
                  },
                ]}
                placeholder='请输入企业微信名'
              />
            </FormItem>
            <FormItem label='完成时间'>
              <DatePicker v-model={timeForm.value} type='daterange' />
            </FormItem>
          </Form>
          <div class='btn-container'>
            <Button theme='primary' onClick={filterOrders}>
              <Search />
              查询
            </Button>
            <Button onClick={clearFilter}>重置</Button>
          </div>
        </div>
        <div class='btn-container oper-btn-pad'>
          <export-to-excel-button data={dataList.value} columns={tableColumns} filename='回收设备列表' />
        </div>
        <CommonTable />
      </div>
    );
  },
});
