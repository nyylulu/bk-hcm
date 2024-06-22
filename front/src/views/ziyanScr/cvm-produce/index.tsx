import { defineComponent, ref, computed, onMounted, onUnmounted } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getRequireTypes } from '@/api/host/task';
import { getDeviceTypes, getCvmProduceOrderList, getCvmProducedResources } from '@/api/host/cvm';
import MemberSelect from '@/components/MemberSelect';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import FastCvmProduce from './component/fast-cvm-produce';
import CreateOrder from './component/create-order';
import SuccessProduceDetail from './component/success-produce-detail';
import { Button, Form, Select, TagInput } from 'bkui-vue';
import { Search } from 'bkui-vue/lib/icon';
import { statusList } from './transform';
import { merge, throttle } from 'lodash';
import dayjs from 'dayjs';
import './index.scss';
const { FormItem } = Form;
export default defineComponent({
  components: {
    MemberSelect,
    AreaSelector,
    ZoneSelector,
    FastCvmProduce,
    CreateOrder,
    SuccessProduceDetail,
  },
  setup() {
    const defaultCvmProduceForm = () => ({
      require_type: [],
      region: [],
      zone: [],
      device_type: [],
      order_id: [],
      task_id: [],
      status: [],
      bk_username: [],
    });
    const defaultTime = () => [new Date(dayjs().subtract(1, 'week').format('YYYY-MM-DD')), new Date()];
    const cvmProduceForm = ref(defaultCvmProduceForm());
    const timeForm = ref(defaultTime());
    const handleTime = (time) => (!time ? '' : dayjs(time).format('YYYY-MM-DD'));
    const timeObj = computed(() => {
      return {
        start: handleTime(timeForm.value[0]),
        end: handleTime(timeForm.value[1]),
      };
    });
    const pageInfo = ref({
      start: 0,
      limit: 10,
    });
    const requestListParams = ref({
      ...timeObj.value,
      page: pageInfo.value,
    });
    const loadOrders = () => {
      const params = {
        ...cvmProduceForm.value,
        ...timeObj.value,
        page: pageInfo.value,
      };
      params.order_id = params.order_id.map((item) => Number(item));
      requestListParams.value = { ...params };
      getListData();
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      loadOrders();
    };
    const clearFilter = () => {
      cvmProduceForm.value = defaultCvmProduceForm();
      timeForm.value = defaultTime();
      filterOrders();
    };
    const orderClipboard = ref({});
    const isShowProduceDetail = ref(false);
    const handleCheckSuccessNum = () => {
      isShowProduceDetail.value = true;
    };
    const { columns } = useColumns('cvmProduceQuery');
    columns.splice(9, 0, {
      label: '生产情况-成功',
      field: 'success_num',
      width: 150,
      render: ({ row }) => {
        if (row.success_num > 0) {
          const { order_id } = row;
          const ips = orderClipboard.value[order_id]?.ips || [];
          const assetIds = orderClipboard.value[order_id]?.assetIds || [];

          return (
            <div class='success-container'>
              <div>
                <Button text theme='primary' onClick={handleCheckSuccessNum}>
                  {row.success_num}
                </Button>
              </div>
              <div v-bk-tooltips={{ placement: 'top', content: '复制 IP' }}>
                <Button text theme='primary' v-clipboard={ips.join('\n')}>
                  复制 IP
                </Button>
              </div>
              <div v-bk-tooltips={{ placement: 'top', content: '复制固资号' }}>
                <Button text theme='primary' v-clipboard={assetIds.join('\n')}>
                  复制固资号
                </Button>
              </div>
            </div>
          );
        }
        return <span>{row.success_num}</span>;
      },
    });
    const tableColumns = [...columns];
    const { CommonTable, getListData, dataList, pagination, sort, order } = useTable({
      tableOptions: {
        columns: tableColumns,
        extra: {
          onRowMouseEnter: (e, row) => {
            handleCellMouseEnter(row);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/cvm/findmany/apply/order',
          payload: {
            ...requestListParams.value,
          },
        };
      },
    });
    const isCreateOrderVisible = ref(false);
    const fastProduceData = ref({});
    const handleOrderCreate = (resource) => {
      isCreateOrderVisible.value = true;
      if (resource) {
        fastProduceData.value = resource;
      }
    };
    const clearDataInfo = () => {
      fastProduceData.value = {};
    };
    const isFastProVisible = ref(false);
    const handleFastCvmProduce = () => {
      isFastProVisible.value = true;
    };
    // 需求类型
    const requireTypeList = ref([]);
    const fetchRequireType = async () => {
      const res = await getRequireTypes();
      requireTypeList.value = res.data.info.map((item) => ({
        label: item.require_name,
        value: item.require_type,
      }));
    };
    // CVM机型
    const cvmDeviceTypeList = ref([]);
    const fetchCvmDeviceType = async () => {
      const res = await getDeviceTypes({});
      cvmDeviceTypeList.value = res.data.info.map((item) => ({
        label: item,
        value: item,
      }));
    };
    const queryProduceOrder = () => {
      pageInfo.value.start = 0;
      loadOrders(true);
    };
    const updateCvmProduce = () => {
      queryProduceOrder();
    };
    const throttleLoadHostInfo = ref(null);
    const loadProducedResources = (orderId) => {
      return getCvmProducedResources({ order_id: orderId });
    };
    const producedDetail = ref([]);
    const loadCvmProduceDetail = () => {
      throttleLoadHostInfo.value = throttle(
        async (row) => {
          // const [, res] = await to(this.loadProducedResources(row.order_id));
          const res = await loadProducedResources(row.order_id);
          const ips = res.data.info.map((item) => item.ip);
          const assetIds = res.data.info.map((item) => item.asset_id);
          orderClipboard.value[row.order_id] = {
            ips,
            assetIds,
          };
          producedDetail.value = res?.data?.info || [];
        },
        500,
        { trailing: true },
      );
    };
    const handleCellMouseEnter = (row) => {
      if (row.success_num > 0) {
        throttleLoadHostInfo.value(row);
      }
    };
    const poller = ref(null);
    const pollProduceOrderList = () => {
      const newPage = {
        start: pagination.start,
        limit: pagination.limit,
        sort: `${sort.value}:${order.value === 'ASC' ? 1 : -1}`,
      };
      const params = Object.assign(requestListParams.value, { page: newPage });
      getCvmProduceOrderList(params).then((res) => {
        dataList.value.forEach((currentOrder) => {
          const newOrder = res?.data?.info?.find((item) => item.order_id === currentOrder.order_id) || null;
          if (newOrder) {
            merge(currentOrder, newOrder);
          }
        });
      });
    };
    onMounted(() => {
      fetchRequireType();
      fetchCvmDeviceType();
      loadCvmProduceDetail();
      if (poller.value) clearInterval(poller.value);
      poller.value = setInterval(() => {
        pollProduceOrderList();
      }, 5000);
    });
    onUnmounted(() => {
      clearInterval(poller.value);
    });
    return () => (
      <>
        <CommonTable>
          {{
            expandRow: (row) => {
              return (
                <property-list
                  properties={{
                    imageId: row.spec.image_id,
                    diskType: row.spec.disk_type,
                    diskSize: row.spec.disk_size,
                    bkBizId: row.bk_biz_id,
                    module: 'SA云化池',
                    vpc: row.spec.vpc,
                    subnet: row.spec.subnet,
                  }}
                />
              );
            },
            tabselect: () => (
              <Form label-width='110' class='cvm-produce-form' model={cvmProduceForm}>
                <FormItem label-width='0'>
                  <Button theme='primary' onClick={() => handleOrderCreate(false)}>
                    创建单据
                  </Button>
                </FormItem>
                <FormItem label-width='0'>
                  <Button theme='primary' onClick={handleFastCvmProduce}>
                    快速生产
                  </Button>
                </FormItem>
                <FormItem label='需求类型'>
                  <Select v-model={cvmProduceForm.value.require_type} multiple clearable placeholder='请选择'>
                    {requireTypeList.value.map(({ label, value }) => {
                      return <Select.Option key={value} name={label} id={value} />;
                    })}
                  </Select>
                </FormItem>
                <FormItem label='地域'>
                  <area-selector
                    multiple
                    v-model={cvmProduceForm.value.region}
                    params={{ resourceType: 'QCLOUDCVM' }}
                  />
                </FormItem>
                <FormItem label='园区'>
                  <zone-selector
                    multiple
                    v-model={cvmProduceForm.value.zone}
                    params={{ resourceType: 'QCLOUDCVM', region: cvmProduceForm.value.region }}
                  />
                </FormItem>
                <FormItem label='机型'>
                  <Select v-model={cvmProduceForm.value.device_type} multiple clearable placeholder='请选择'>
                    {cvmDeviceTypeList.value.map(({ value, label }) => {
                      return <Select.Option key={value} name={label} id={value} />;
                    })}
                  </Select>
                </FormItem>
                <FormItem label='单号'>
                  <TagInput
                    class='tag-input-width'
                    v-model={cvmProduceForm.value.order_id}
                    placeholder='请输入单号'
                    allow-create
                    has-delete-icon
                    // allow-auto-match
                  />
                </FormItem>
                <FormItem label='云梯单号'>
                  <TagInput
                    class='tag-input-width'
                    v-model={cvmProduceForm.value.task_id}
                    placeholder='请输入云梯单号'
                    allow-create
                    has-delete-icon
                  />
                </FormItem>
                <FormItem label='状态'>
                  <Select v-model={cvmProduceForm.value.status} multiple clearable placeholder='请选择状态'>
                    {statusList.value.map(({ status, description }) => {
                      return <Select.Option key={status} name={description} id={status} />;
                    })}
                  </Select>
                </FormItem>
                <FormItem label='创建人'>
                  <member-select
                    class='tag-input-width'
                    v-model={cvmProduceForm.value.bk_username}
                    multiple
                    clearable
                    placeholder='请输入企业微信名'
                  />
                </FormItem>
                <FormItem label='回收时间'>
                  <bk-date-picker v-model={timeForm.value} type='daterange' />
                </FormItem>
                <FormItem label-width='0' class='cvm-produce-form-btn'>
                  <Button theme='primary' onClick={filterOrders}>
                    <Search />
                    查询
                  </Button>
                  <Button onClick={() => clearFilter()}>清空</Button>
                </FormItem>
              </Form>
            ),
          }}
        </CommonTable>
        <create-order
          v-model={isCreateOrderVisible.value}
          onUpdateProduceData={updateCvmProduce}
          onClearDataInfo={clearDataInfo}
          dataInfo={fastProduceData.value}
        />
        <fast-cvm-produce v-model={isFastProVisible.value} onOneKeyApply={handleOrderCreate} />
        <success-produce-detail v-model={isShowProduceDetail.value} tableData={producedDetail.value} />
      </>
    );
  },
});
