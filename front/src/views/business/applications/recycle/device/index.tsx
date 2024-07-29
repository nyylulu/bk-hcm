import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { DatePicker, Select } from 'bkui-vue';
import GridFilterComp from '@/components/grid-filter-comp';
import ExportToExcelButton from '@/components/export-to-excel-button';
import FloatInput from '@/components/float-input';
import MemberSelect from '@/components/MemberSelect';

import dayjs from 'dayjs';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '@/store';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { getDeviceTypeList, getRecycleStageOpts, getRegionList, getZoneList } from '@/api/host/recycle';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export default defineComponent({
  setup() {
    const router = useRouter();
    const userStore = useUserStore();
    const { t } = useI18n();
    const { getBusinessApiPath, getBizsId } = useWhereAmI();

    const defaultDeviceForm = () => ({
      bk_biz_id: [] as number[],
      order_id: [] as any[],
      suborder_id: [] as any[],
      ip: [] as any[],
      device_type: [] as any[],
      bk_zone_name: [] as any[],
      sub_zone: [] as any[],
      stage: [] as any[],
      bk_username: [userStore.username],
    });
    const defaultTime = () => [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()];

    const deviceForm = ref(defaultDeviceForm());
    const timeForm = ref(defaultTime());
    const pageInfo = ref({ start: 0, limit: 10, enable_count: false });

    const deviceTypeList = ref([]);
    const bkZoneNameList = ref([]);
    const subZoneList = ref([]);
    const stageList = ref([]);

    const handleTime = (time: any) => (!time ? '' : dayjs(time).format('YYYY-MM-DD'));
    const timeObj = computed(() => {
      return {
        start: handleTime(timeForm.value[0]),
        end: handleTime(timeForm.value[1]),
      };
    });
    const requestListParams = computed(() => {
      const params = {
        ...deviceForm.value,
        ...timeObj.value,
        page: pageInfo.value,
        bk_biz_id: [getBizsId()],
      };
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      removeEmptyFields(params);
      return params;
    });

    const { columns } = useScrColumns('deviceQuery');
    columns.splice(1, 0, {
      label: t('子单号'),
      field: 'suborder_id',
      width: 80,
      render: ({ row }: any) => {
        return (
          // 单据详情
          <span class='sub-order-num' onClick={() => enterDetail(row)}>
            {row.suborder_id}
          </span>
        );
      },
    });
    const { CommonTable, getListData, dataList, pagination, isLoading } = useTable({
      tableOptions: {
        columns,
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: { sort: 'ip', order: 'ASC' },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/host`,
          payload: {
            ...requestListParams.value,
          },
        };
      },
    });
    const enterDetail = (row: any) => {
      router.push({ name: 'HostRecycleDocDetail', query: { suborderId: row.suborder_id, bkBizId: getBizsId() } });
    };

    const filterOrders = () => {
      pagination.start = 0;
      deviceForm.value.bk_biz_id = [getBizsId()];
      getListData();
    };
    const clearFilter = () => {
      const initForm = defaultDeviceForm();
      initForm.bk_biz_id = [getBizsId()];
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

    return () => (
      <>
        <GridFilterComp
          rules={[
            {
              title: t('单号'),
              content: <FloatInput v-model={deviceForm.value.order_id} placeholder={t('请输入单号，多个换行分割')} />,
            },
            {
              title: t('子单号'),
              content: (
                <FloatInput v-model={deviceForm.value.suborder_id} placeholder={t('请输入子单号，多个换行分割')} />
              ),
            },
            {
              title: t('机型'),
              content: (
                <Select v-model={deviceForm.value.device_type} multiple clearable placeholder={t('请选择机型')}>
                  {deviceTypeList.value.map((item) => {
                    return <Select.Option key={item} name={item} id={item} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('地域'),
              content: (
                <Select v-model={deviceForm.value.bk_zone_name} multiple clearable placeholder={t('请选择地域')}>
                  {bkZoneNameList.value.map((item) => {
                    return <Select.Option key={item} name={item} id={item} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('园区'),
              content: (
                <Select v-model={deviceForm.value.sub_zone} multiple clearable placeholder={t('请选择园区')}>
                  {subZoneList.value.map((item) => {
                    return <Select.Option key={item} name={item} id={item} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('状态'),
              content: (
                <Select v-model={deviceForm.value.stage} multiple clearable placeholder={t('请选择状态')}>
                  {stageList.value.map(({ stage, description }) => {
                    return <Select.Option key={stage} name={description} id={stage} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('回收IP'),
              content: <FloatInput v-model={deviceForm.value.ip} placeholder={t('请输入IP，多个换行分割')} />,
            },
            {
              title: t('回收人'),
              content: (
                <MemberSelect
                  v-model={deviceForm.value.bk_username}
                  multiple
                  clearable
                  placeholder={t('请输入企业微信名')}
                  defaultUserlist={[{ username: userStore.username, display_name: userStore.username }]}
                />
              ),
            },
            {
              title: t('完成时间'),
              content: <DatePicker class='full-width' v-model={timeForm.value} type='daterange' />,
            },
          ]}
          onSearch={filterOrders}
          onReset={clearFilter}
          loading={isLoading.value}
          col={5}
          immediate
        />
        <section class={cssModule['table-wrapper']}>
          <div class={[cssModule.buttons, cssModule.mb16]}>
            <ExportToExcelButton
              class={cssModule.button}
              data={dataList.value}
              columns={columns}
              text={t('全部导出')}
              filename={t('回收设备列表')}
            />
          </div>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </>
    );
  },
});
