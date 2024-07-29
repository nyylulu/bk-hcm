import { defineComponent, ref, computed, onMounted } from 'vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { exportTableToExcel, getEntirePath } from '@/utils';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import { Button, Form } from 'bkui-vue';
import { useRoute } from 'vue-router';
import ExecuteRecord from '../execute-record';
import FloatInput from '@/components/float-input';
import './index.scss';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';

const { FormItem } = Form;

export default defineComponent({
  components: {
    ExecuteRecord,
    FloatInput,
  },
  setup() {
    const route = useRoute();
    const { getBusinessApiPath } = useWhereAmI();
    const defaultFilter = () => ({
      order_id: [],
      suborder_id: route?.query?.suborder_id?.split('\n') || [],
      ip: [],
    });
    const filter = ref(defaultFilter());
    const page = ref({
      limit: 10,
      start: 0,
    });
    const requestParams = computed(() => {
      return (data) => {
        const params = {
          ...data,
          page,
        };
        params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
        removeEmptyFields(params);
        return params;
      };
    });
    const getDyParams = ref(requestParams.value(filter.value));
    const querying = ref(false);
    const { columns } = useColumns('pdExecutecolumns');
    const PDcolumns = [...columns];
    PDcolumns.splice(2, 0, {
      label: 'IP',
      field: 'ip',
      render: ({ row }) => {
        return (
          <Button
            text
            theme='primary'
            disabled={row.listener_num > 0 || row.delete_protect}
            onClick={() => application(row)}>
            {row.ip}
          </Button>
        );
      },
    });
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: [
          ...PDcolumns,
          {
            label: '操作',
            width: 120,
            render: ({ data }: { data: any }) => {
              return (
                <Button
                  text
                  theme='primary'
                  disabled={data.listener_num > 0 || data.delete_protect}
                  onClick={() => application(data)}>
                  详情
                </Button>
              );
            },
          },
        ],
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
          payload: {
            ...getDyParams.value,
          },
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/detect`,
        };
      },
    });
    const getPdList = () => {
      const params = {
        ...filter.value,
      };
      getDyParams.value = requestParams.value(params);
      getListData();
      getIpList();
    };
    const failIpList = ref([]);
    const allIpList = ref([]);
    const getIpList = async () => {
      const params = {
        ...requestParams.value(filter.value),
        page: { start: 0, limit: 500 },
      };
      const [failIpData, allIpData] = await Promise.all([
        http
          .post(getEntirePath(`${getBusinessApiPath()}task/list/recycle/detect/host`), {
            ...params,
            status: ['FAILED'],
          })
          .then((res: any) => res.data),
        http
          .post(getEntirePath(`${getBusinessApiPath()}task/list/recycle/detect/host`), params)
          .then((res: any) => res.data),
      ]);
      failIpList.value = failIpData?.info || [];
      allIpList.value = allIpData?.info || [];
    };
    const filterOrders = () => {
      page.value.start = 0;
      getPdList();
    };
    const clearFilter = () => {
      filter.value = defaultFilter();
      filterOrders();
    };
    const exportToExcel = async () => {
      const res = await http.post(
        getEntirePath(`${getBusinessApiPath()}task/findmany/recycle/detect`),
        Object.assign(requestParams.value(filter.value), { page: { start: 0, limit: 500 } }),
      );

      const totalList = res.data?.info || [];
      exportTableToExcel(totalList, [{ label: 'IP', field: 'ip', width: 80 }, ...columns], '预检详情列表');
    };
    const openDetails = ref(false);
    const transferData = ref({});
    const application = (row) => {
      openDetails.value = true;
      Object.assign(transferData.value, {
        suborderId: row.suborder_id,
        ip: row.ip,
        page: {
          start: 0,
          limit: 10,
        },
      });
    };
    onMounted(() => {
      getIpList();
    });
    return () => (
      <div class={'application-detail-container'}>
        <DetailHeader>预检详情</DetailHeader>
        <div class={'detail-wrapper'}>
          <div class={'apply-list-container'}>
            <div class={'filter-container'}>
              <Form model={filter.value} formType='vertical' class={'scr-form-wrapper'}>
                <FormItem label='单号'>
                  <FloatInput v-model={filter.value.order_id} placeholder='请输入单号，多个换行分割' />
                </FormItem>
                <FormItem label='子单号'>
                  <FloatInput v-model={filter.value.suborder_id} placeholder='请输入子单号，多个换行分割' />
                </FormItem>
                <FormItem label='IP'>
                  <FloatInput v-model={filter.value.ip} placeholder='请输入IP，多个换行分割' />
                </FormItem>
              </Form>
              <div class='btn-container'>
                <Button theme='primary' onClick={filterOrders} loading={querying.value}>
                  <Search />
                  查询
                </Button>
                <Button onClick={clearFilter}>重置</Button>
              </div>
            </div>
            <div class='btn-container oper-btn-pad'>
              <Button disabled={allIpList.value.length === 0} v-clipboard={allIpList.value.join('\n')}>
                复制所有主机IP <span>({allIpList.value.length})</span>
              </Button>
              <Button disabled={failIpList.value.length === 0} v-clipboard={failIpList.value.join('\n')}>
                复制失败主机IP <span>({failIpList.value.length})</span>
              </Button>
              <Button onClick={exportToExcel}>导出全部</Button>
            </div>
          </div>
          <CommonTable />
          <execute-record v-model={openDetails.value} dataInfo={transferData.value} />
        </div>
      </div>
    );
  },
});
