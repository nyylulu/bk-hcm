import { computed, defineComponent, ref, watch, onMounted } from 'vue';
import { RouteLocationRaw, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button, Message, TagInput } from 'bkui-vue';
import { Copy, DataShape, HelpDocumentFill } from 'bkui-vue/lib/icon';
import ScrDatePicker from '@/components/scr/scr-date-picker';
import ScrCreateFilterSelector from '@/views/ziyanScr/resource-manage/create/ScrCreateFilterSelector';
import MemberSelect from '@/components/MemberSelect';
import WName from '@/components/w-name';
import StageDetailSideslider from './stage-detail';

import moment from 'moment';
import { useI18n } from 'vue-i18n';
import { throttle } from 'lodash';
import { useAccountStore, useUserStore, useZiyanScrStore } from '@/store';
import useFormModel from '@/hooks/useFormModel';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import { applicationTime } from '@/common/util';
import { getResourceTypeName } from '@/views/ziyanScr/hostApplication/components/transform';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import http from '@/http';
import GridFilterComp from '@/components/grid-filter-comp';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const router = useRouter();
    const { t } = useI18n();
    const accountStore = useAccountStore();
    const userStore = useUserStore();
    const scrStore = useZiyanScrStore();

    const { transformApplyStages } = useApplyStages();
    const { transformRequireTypes } = useRequireTypes();

    const stageDetailSidesliderRef = ref();

    const { formModel, resetForm } = useFormModel({
      bkBizId: [],
      requireType: [],
      stage: [],
      orderId: [],
      dateRange: applicationTime(),
      user: [userStore.username],
    });
    const curRow = ref({});

    const isMatchPanelShow = ref(false);
    const orderClipboard = ref<any>({});
    const machineDetails = ref([]);

    const { columns } = useScrColumns('applicationList');
    columns.splice(3, 0);

    const opBtnDisabled = computed(() => {
      return (row: any) => {
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

    const getOrderRoute = (row: any) => {
      let routeParams: RouteLocationRaw = {
        name: 'HostApplicationsDetail',
        params: { id: row.order_id },
      };
      if (row.stage === 'UNCOMMIT') {
        // todo: 需要更换指业务下主机申请的路由
        routeParams = { name: '提交主机申请', query: { order_id: row.order_id, unsubmitted: 1 } };
      }
      router.push(routeParams);
    };

    const modify = (data: any) => {
      router.push({ name: 'HostApplicationsModify', query: { ...data } });
    };

    const reapply = (data: any) => {
      // todo: 需要更换指业务下主机申请的路由
      router.push({ name: '提交主机申请', query: { order_id: data.order_id, unsubmitted: 0 } });
    };

    const throttleInfo = ref(null);
    // 已交付设备
    const getDeliveredDevices = (params: any) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/apply/device`, params);
    };
    // 查询交付IP和固号IP
    const getDeliveredHostField = (row: any, fieldKey: any) => {
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
      return getDeliveredDevices(params).then((res: any) => {
        const value = res?.data?.info?.map((item: any) => item[fieldKey]) || [];
        return value;
      });
    };
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
    const handleCellMouseEnter = (row: any) => {
      if (row.success_num > 0) {
        throttleInfo.value(row);
      }
    };
    // 获取匹配详情
    const getMatchDetails = async (subOrderId: number) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/apply/detail`, {
        suborder_id: subOrderId,
      });
    };

    const { CommonTable, getListData, isLoading, pagination } = useTable({
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
            label: '单据状态',
            field: 'stage',
            width: 200,
            render: ({ data }: any) => {
              const { stage, createAt, modify_time: modifyTime } = data;
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
                            {modifyTime < 2 ? (
                              <span>
                                建议
                                <Button size='small' text theme={'primary'} class={'ml8'}>
                                  修改需求重试
                                </Button>
                              </span>
                            ) : (
                              <span>
                                请查看详情后联系 <WName name={'BK助手'} class={'ml8'}></WName> 进行处理
                              </span>
                            )}
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
                    class={{ ml8: stage === 'SUSPEND' && modifyTime < 2 }}
                    onClick={async () => {
                      const { data: list } = await getMatchDetails(data.suborder_id);
                      machineDetails.value = list.info;
                      stageDetailSidesliderRef.value.triggerShow(true);
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
                    {stage === 'SUSPEND' && modifyTime < 2 ? modifyButton() : null}
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
            render: ({ data }: any) => transformRequireTypes(data.requireType),
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
            field: 'total_num',
          },
          {
            label: '待交付数',
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
            width: 120,
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
                    disabled={data.status === 'UNCOMMIT'}
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
                      Message({ theme: 'success', message: '重试成功' });
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
                      Message({ theme: 'success', message: '终止成功' });
                    }}>
                    终止
                  </Button>
                </div>
              );
            },
          },
        ],
        extra: {
          onRowMouseEnter: (e: any, row: any) => {
            handleCellMouseEnter(row);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        immediate: false,
      },
      scrConfig: () => ({
        url: '/api/v1/woa/task/findmany/apply',
        payload: removeEmptyFields({
          bk_biz_id: [accountStore.bizs],
          order_id: formModel.orderId.map((v) => Number(v)),
          bk_username: formModel.user,
          stage: formModel.stage,
          start: formModel.dateRange[0],
          end: formModel.dateRange[1],
          require_type: formModel.requireType,
        }),
      }),
    });

    const filterOrders = () => {
      pagination.start = 0;
      formModel.bkBizId = [accountStore.bizs];
      getListData();
    };

    onMounted(() => {
      throttleDeliveredHostField();
    });

    watch(
      () => userStore.username,
      (username) => {
        formModel.user = [username];
      },
    );

    return () => (
      <>
        <GridFilterComp
          rules={[
            {
              title: t('需求类型'),
              content: (
                <ScrCreateFilterSelector
                  v-model={formModel.requireType}
                  api={scrStore.getRequirementList}
                  multiple
                  optionIdPath='require_type'
                  optionNamePath='require_name'
                />
              ),
            },
            {
              title: t('单据状态'),
              content: (
                <ScrCreateFilterSelector
                  v-model={formModel.stage}
                  api={scrStore.getApplyStageList}
                  multiple
                  optionIdPath='stage'
                  optionNamePath='description'
                />
              ),
            },
            {
              title: t('单号'),
              content: (
                <TagInput
                  v-model={formModel.orderId}
                  allow-create
                  collapse-tags
                  allow-auto-match
                  pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
                  placeholder='请输入单号'
                />
              ),
            },
            {
              title: t('申请时间'),
              content: <ScrDatePicker class='full-width' v-model={formModel.dateRange} />,
            },
            {
              title: t('申请人'),
              content: (
                <MemberSelect
                  v-model={formModel.user}
                  clearable
                  defaultUserlist={[
                    {
                      username: userStore.username,
                      display_name: userStore.username,
                    },
                  ]}
                />
              ),
            },
          ]}
          onSearch={filterOrders}
          onReset={() => {
            resetForm({ user: [userStore.username] });
            formModel.bkBizId = [accountStore.bizs];
            filterOrders();
          }}
          loading={isLoading.value}
        />
        <section class={cssModule['table-wrapper']}>
          <CommonTable />
        </section>
        <StageDetailSideslider ref={stageDetailSidesliderRef} details={machineDetails.value} />
      </>
    );
  },
});
