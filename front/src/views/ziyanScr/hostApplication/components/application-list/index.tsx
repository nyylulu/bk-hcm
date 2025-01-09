import { defineComponent, onMounted, ref, computed, watch, reactive } from 'vue';
import './index.scss';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { Button, Message, Table, Sideslider } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useRoute, useRouter } from 'vue-router';
import moment from 'moment';
import WName from '@/components/w-name';
import { Copy, DataShape, HelpDocumentFill } from 'bkui-vue/lib/icon';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import CommonSideslider from '@/components/common-sideslider';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import { timeFormatter } from '@/common/util';
import http from '@/http';
import { useZiyanScrStore } from '@/store';
import SuborderDetail from '../suborder-detail';
import CommonDialog from '@/components/common-dialog';
import { throttle } from 'lodash';
import MatchPanel from '../match-panel';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getResourceTypeName } from '../transform';
import { getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import useSearchQs from '@/hooks/use-search-qs';
import { useBusinessGlobalStore } from '@/store/business-global';
import { getDateRange, transformFlatCondition } from '@/utils/search';
import type { ModelProperty } from '@/model/typings';
import { getModel } from '@/model/manager';
import HocSearch from '@/model/hoc-search.vue';
import { HostApplySearchNonBusiness } from '@/model/order/host-apply-search';
import { serviceShareBizSelectedKey } from '@/constants/storage-symbols';

export default defineComponent({
  setup() {
    const businessMapStore = useBusinessMapStore();
    const { transformApplyStages } = useApplyStages();
    const machineDetails = ref([]);
    const isMatchPanelShow = ref(false);
    const isDialogShow = ref(false);
    const curRow = ref({});
    const curSuborder = ref({
      step_name: '',
      step_id: 1,
      suborder_id: 0,
    });

    const stageDetailSlideState = reactive({
      show: false,
      suborderId: undefined,
    });

    const scrStore = useZiyanScrStore();
    const businessGlobalStore = useBusinessGlobalStore();

    const reapply = (data: any) => {
      router.push({
        path: '/service/hostApplication/apply',
        query: { order_id: data.order_id, unsubmitted: 0 },
      });
    };
    const modify = (data: any) => {
      router.push({
        path: '/service/hostApplication/modify',
        query: { ...data },
      });
    };

    const { columns } = useColumns('applicationList');
    const router = useRouter();
    const route = useRoute();
    const orderClipboard = ref({});
    columns.splice(3, 0);
    const opBtnDisabled = computed(() => {
      return (row) => {
        if (row.stage === 'RUNNING' && row.status === 'MATCHING') {
          return true;
        }
        if (!row.suborder_id) {
          return true;
        }
        if (
          ['wait', 'MATCHED_SOME', 'MATCHING'].includes(row.status) ||
          (row.stage === 'SUSPEND' && row.status === 'TERMINATE')
        ) {
          return false;
        }
        if (['UNCOMMIT', 'PAUSED'].includes(row.status)) {
          return true;
        }
        if (['AUDIT'].includes(row.stage) && !row.status) {
          return true;
        }
        if (['TERMINATE'].includes(row.stage)) {
          return true;
        }
        if (row.stage === 'DONE' && row.status === 'DONE') {
          return true;
        }
        return false;
      };
    });

    const searchFields = getModel(HostApplySearchNonBusiness).getProperties();
    const searchQs = useSearchQs({ key: 'filter', properties: searchFields });

    const { CommonTable, getListData, pagination } = useTable({
      tableOptions: {
        columns: [
          {
            label: '单号/子单号',
            width: 100,
            render: ({ data }: any) => {
              return (
                <div>
                  <div>
                    <Button theme='primary' text onClick={() => getOrderRoute(data)}>
                      {data.order_id}
                    </Button>
                  </div>
                  <div>
                    <p>{data.suborder_id || '无'}</p>
                  </div>
                </div>
              );
            },
          },
          {
            label: '业务',
            render: ({ data }: any) =>
              businessMapStore.getNameFromBusinessMap(data.bk_biz_id) || data.bk_biz_id || '--',
          },
          {
            label: '单据状态',
            field: 'stage',
            width: 200,
            render: ({ data }: any) => {
              const { stage, createAt } = data;
              const diffHours = moment(new Date()).diff(moment(createAt), 'hours');
              const isAbnormal = diffHours >= 2 && stage === 'RUNNING';

              const stageClass = (stage: string) => {
                if (stage === 'UNCOMMIT') return 'c-text-3';
                if (stage === 'AUDIT') return 'c-text-2';
                if (stage === 'DONE') return 'c-success';
                if (isAbnormal) return 'c-warning';
                if (stage === 'RUNNING') return 'c-text-1';
                if (stage === 'TERMINATE') return 'c-danger';
                if (stage === 'SUSPEND') return 'c-danger';
              };

              const abnormalStatus = () => {
                if (stage === 'SUSPEND') {
                  return (
                    <div
                      class={'flex-row align-item-center'}
                      v-bk-tooltips={{
                        content: (
                          <span>
                            建议
                            <Button size='small' text theme={'primary'} class={'ml8'}>
                              修改需求重试
                            </Button>
                          </span>
                        ),
                      }}>
                      备货状态异常 <HelpDocumentFill fill='#ffbb00' width={12} height={12} class={'ml4'} />
                    </div>
                  );
                }
                return null;
              };

              const modifyButton = () => {
                return (
                  <Button
                    class='mr8'
                    size='small'
                    onClick={() => modify(data)}
                    disabled={data.resource_type === 'IDCPM'}
                    v-bk-tooltips={{
                      content: 'IDC物理机不支持修改,请联系ICR(IEG资源服务助手)',
                      disabled: data.resource_type !== 'IDCPM',
                    }}
                    text
                    theme={'primary'}>
                    修改需求重试
                  </Button>
                );
              };

              const progressButton = () => {
                return (
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    onClick={async () => {
                      stageDetailSlideState.show = true;
                      stageDetailSlideState.suborderId = data.suborder_id;
                      const { data: list } = await getMatchDetails(data.suborder_id);
                      machineDetails.value = list.info;
                    }}>
                    查看详情
                  </Button>
                );
              };

              return (
                <div>
                  <p class={stageClass(stage)}>
                    {stage !== 'SUSPEND' && transformApplyStages(stage)}
                    {abnormalStatus()}
                  </p>
                  <p>
                    {stage === 'SUSPEND' ? modifyButton() : null}
                    {['RUNNING', 'DONE', 'SUSPEND'].includes(stage) ? progressButton() : null}
                  </p>
                </div>
              );
            },
          },
          {
            label: '需求类型',
            field: 'require_type',
            width: 100,
            render: ({ data }: any) => getTypeCn(data.require_type),
          },
          {
            label: '需求摘要',
            width: 250,
            render: ({ data }: any) => {
              return (
                <div>
                  <div style={'height: 30px!important;line-height: 30px;'}>
                    资源类型：{getResourceTypeName(data?.resource_type)}
                  </div>
                  <div style={'height: 20px!important;line-height: 20px;'}>机型：{data.spec?.device_type || '--'}</div>
                  <div style={'height: 30px!important;line-height: 30px;'}>园区：{getZoneCn(data.spec?.zone)}</div>
                </div>
              );
            },
          },
          {
            label: '申请人',
            render: ({ data }: any) => {
              return <WName name={data.bk_username}></WName>;
            },
          },
          {
            label: `需求数`,
            width: 90,
            field: 'total_num',
          },
          {
            label: '待交付数',
            width: 90,
            field: 'pending_num',
            render({ cell, data }: any) {
              return cell ? (
                <Button
                  theme='primary'
                  text
                  onClick={() => {
                    curRow.value = data;
                    isMatchPanelShow.value = true;
                  }}>
                  {cell}
                </Button>
              ) : (
                cell
              );
            },
          },
          {
            label: '已交付数',
            field: 'success_num',
            width: 180,
            render: ({ data }: any) => {
              if (data.success_num > 0) {
                const ips = orderClipboard.value?.[data.suborder_id]?.ips || [];
                const assetIds = orderClipboard.value?.[data.suborder_id]?.assetIds || [];
                const goToCmdb = (ips: string[]) => {
                  window.open(`http://bkcc.oa.com/#/business/${data.bk_biz_id}/index?ip=text=${ips.join(',')}`);
                };

                return (
                  <div class={'flex-row align-item-center'}>
                    {data.success_num}
                    <Button
                      text
                      theme={'primary'}
                      class='ml8 mr8'
                      v-clipboard:copy={ips.join('\n')}
                      v-bk-tooltips={{
                        content: '复制 IP',
                      }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      class='mr8'
                      v-clipboard:copy={assetIds.join('\n')}
                      v-bk-tooltips={{
                        content: '复制固资号',
                      }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      onClick={() => goToCmdb(ips)}
                      v-bk-tooltips={{
                        content: '去蓝鲸配置平台管理资源',
                      }}>
                      <DataShape />
                    </Button>
                  </div>
                );
              }

              return <span>{data.success_num}</span>;
            },
          },
          ...columns,
          {
            label: '操作',
            fixed: 'right',
            width: 200,
            render: ({ data }: any) => {
              return (
                <div>
                  <Button
                    // 滚服项目暂不支持再次申请
                    disabled={data.status === 'UNCOMMIT' || data.require_type === 6}
                    size='small'
                    onClick={() => reapply(data)}
                    text
                    theme={'primary'}
                    class='mr8'>
                    再次申请
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={opBtnDisabled.value(data)}
                    onClick={async () => {
                      await scrStore.retryOrder({ suborder_id: [data.suborder_id] });
                      Message({
                        theme: 'success',
                        message: '重试成功',
                      });
                    }}>
                    重试
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={opBtnDisabled.value(data)}
                    onClick={async () => {
                      await scrStore.stopOrder({ suborder_id: [data.suborder_id] });
                      Message({
                        theme: 'success',
                        message: '终止成功',
                      });
                    }}>
                    终止
                  </Button>
                </div>
              );
            },
          },
        ],
        extra: {
          onRowMouseEnter: (e, row) => {
            handleCellMouseEnter(row);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        immediate: false,
      },
      scrConfig: () => {
        const payload = transformFlatCondition(condition.value, searchFields);
        if (payload.bk_biz_id?.[0] === 0) {
          payload.bk_biz_id = businessGlobalStore.businessAuthorizedList.map((item: any) => item.id);
        }
        return {
          url: '/api/v1/woa/task/findmany/apply',
          payload,
        };
      },
    });

    const condition = ref<Record<string, any>>({});
    const searchValues = ref<Record<string, any>>({});

    const getSearchCompProps = (field: ModelProperty) => {
      if (field.type === 'business') {
        return {
          scope: 'auth',
          showAll: true,
          emptySelectAll: true,
          cacheKey: serviceShareBizSelectedKey,
        };
      }
      if (field.id === 'create_at') {
        return {
          type: 'daterange',
          format: 'yyyy-MM-dd',
        };
      }
      if (field.id === 'order_id') {
        return {
          pasteFn: (value: string) =>
            value
              .split(/\r\n|\n|\r/)
              .filter((tag) => /^\d+$/.test(tag))
              .map((tag) => ({ id: tag, name: tag })),
          placeholder: '请输入单号',
        };
      }
      return {};
    };

    const handleSearch = () => {
      // TODO: 实际无效
      pagination.start = 0;
      searchQs.set(searchValues.value);
    };

    const handleReset = () => {
      searchQs.clear();
    };

    watch(
      () => route.query,
      async (query) => {
        const defaultCondition = {
          create_at: getDateRange('last30d', true),
          bk_biz_id: businessGlobalStore.getCacheSelected(serviceShareBizSelectedKey) ?? [0],
        };
        condition.value = searchQs.get(query, defaultCondition);

        searchValues.value = condition.value;

        getListData();
      },
      { immediate: true },
    );

    watch(
      () => stageDetailSlideState.show,
      (val) => {
        if (val) {
          stageDetailPolling.resume();
        } else {
          stageDetailPolling.reset();
        }
      },
    );

    const stageDetailPolling = useTimeoutPoll(
      async () => {
        const { data: list } = await getMatchDetails(stageDetailSlideState.suborderId);
        machineDetails.value = list.info;
      },
      30000,
      {
        max: 60,
      },
    );

    const getOrderRoute = (row) => {
      let routeParams: any = {
        name: 'host-application-detail',
        params: {
          id: row.order_id,
        },
        query: { creator: row.bk_username, bkBizId: row.bk_biz_id },
      };
      if (row.stage === 'UNCOMMIT') {
        routeParams = {
          path: '/service/hostApplication/apply',
          query: { order_id: row.order_id, unsubmitted: 1 },
        };
      }
      router.push(routeParams);
    };
    // 获取匹配详情
    const getMatchDetails = async (subOrderId: number) => {
      return http.post('/api/v1/woa/task/find/apply/detail', {
        suborder_id: subOrderId,
      });
    };
    // 已交付设备
    const getDeliveredDevices = (params) => {
      return http.post('/api/v1/woa/task/findmany/apply/device', params);
    };
    // 查询交付IP和固号IP
    const getDeliveredHostField = (row, fieldKey) => {
      const params = {
        filter: {
          condition: 'AND',
          rules: [
            {
              field: 'suborder_id',
              operator: 'equal',
              value: row.suborder_id,
            },
            {
              field: 'bk_biz_id',
              operator: 'in',
              value: [row.bk_biz_id],
            },
          ],
        },
      };
      return getDeliveredDevices(params).then((res) => {
        const value = res?.data?.info?.map((item) => item[fieldKey]) || [];
        return value;
      });
    };
    const throttleInfo = ref(null);
    const throttleDeliveredHostField = () => {
      throttleInfo.value = throttle(async (row) => {
        const [ips, assetIds] = await Promise.all([
          getDeliveredHostField(row, 'ip'),
          getDeliveredHostField(row, 'asset_id'),
        ]);
        orderClipboard.value[row.suborder_id] = {
          ips,
          assetIds,
        };
      }, 200);
    };
    const handleCellMouseEnter = (row) => {
      if (row.success_num > 0) {
        throttleInfo.value(row);
      }
    };

    onMounted(() => {
      throttleDeliveredHostField();
    });
    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'} style={{ margin: '0 24px 20px 24px' }}>
          <GridContainer layout='vertical' column={4} content-min-width={300} gap={[16, 60]}>
            {searchFields.map((field) => (
              <GridItemFormElement key={field.id} label={field.name}>
                <HocSearch
                  is={field.type}
                  display={field.meta?.display}
                  v-model={searchValues.value[field.id]}
                  {...getSearchCompProps(field)}
                />
              </GridItemFormElement>
            ))}
            <GridItem span={4}>
              <div style={{ display: 'flex', gap: '8px' }}>
                <bk-button theme='primary' style={{ minWidth: '86px' }} onClick={handleSearch}>
                  查询
                </bk-button>
                <bk-button style={{ minWidth: '86px' }} onClick={handleReset}>
                  重置
                </bk-button>
              </div>
            </GridItem>
          </GridContainer>
        </div>
        <div class='btn-container oper-btn-pad'>
          <Button
            theme='primary'
            onClick={() => {
              router.push({
                path: '/service/hostApplication/apply',
                query: route.query,
              });
            }}>
            新增申请
          </Button>
        </div>
        <CommonTable />
        <CommonSideslider v-model:isShow={stageDetailSlideState.show} title='资源匹配详情' width={1100} noFooter>
          <Table
            showOverflowTooltip
            border={['outer', 'col', 'row']}
            data={machineDetails.value}
            columns={[
              {
                field: 'step_id',
                label: 'ID',
                width: 40,
              },
              {
                field: 'step_name',
                label: '步骤名称',
                width: '100',
              },
              {
                field: 'status',
                label: '状态',
                width: 80,
                render({ data }: any) {
                  if (data.status === -1) return <span>未执行</span>;
                  if (data.status === 0) return <span>成功</span>;
                  if (data.status === 1) return <span>执行中</span>;
                  return <span>失败</span>;
                },
              },
              {
                field: 'message',
                label: '状态说明',
                width: 100,
              },
              {
                label: '概要',
                width: '250',
                render({ data }: any) {
                  return (
                    <div>
                      <span>
                        <span class='c-text-2 fz-12'>总数：</span>
                        <span>{data.total_num || '-'}</span>
                      </span>
                      <span class='ml-10'>
                        <span class='c-text-2 fz-12'>成功：</span>
                        <span class='c-success'>{data.success_num || '-'}</span>
                      </span>
                      <span class='ml-10'>
                        <span class='c-text-2 fz-12'>进行中：</span>
                        <span>{data.running_num || '-'}</span>
                      </span>
                      <span class='ml-10'>
                        <span class='c-text-2 fz-12'>失败：</span>
                        <span class='c-danger'>{data.fail_num || '-'}</span>
                      </span>
                    </div>
                  );
                },
              },
              {
                field: 'start_at',
                label: '开始时间',
                width: 160,
                render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
              },
              {
                field: 'end_at',
                label: '结束时间',
                width: 160,
                render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
              },
              {
                field: 'operation',
                label: '操作',
                render: ({ data }: any) => (
                  <div>
                    {data.step_id > 1 ? (
                      <Button
                        text
                        theme='primary'
                        onClick={() => {
                          isDialogShow.value = true;
                          curSuborder.value = data;
                        }}>
                        查看详情
                      </Button>
                    ) : (
                      '--'
                    )}
                  </div>
                ),
              },
            ]}></Table>
        </CommonSideslider>

        <CommonDialog v-model:isShow={isDialogShow.value} title={`资源${curSuborder.value.step_name}详情`} width={1200}>
          <SuborderDetail
            suborderId={curSuborder.value.suborder_id}
            stepId={curSuborder.value.step_id}
            isShow={isDialogShow.value}
          />
        </CommonDialog>

        <Sideslider v-model:isShow={isMatchPanelShow.value} title='待匹配' width={1600} renderDirective='if'>
          <MatchPanel data={curRow.value} handleClose={() => (isMatchPanelShow.value = false)} />
        </Sideslider>
      </div>
    );
  },
});
