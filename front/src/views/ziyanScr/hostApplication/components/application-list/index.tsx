import { defineComponent, onMounted, ref, computed } from 'vue';
import './index.scss';
import useFormModel from '@/hooks/useFormModel';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { Button, Form, Input, Message, Table } from 'bkui-vue';
import BusinessSelector from '@/components/business-selector/index.vue';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import ApplicationStatusSelector from '@/components/scr/application-status-selector';
import ScrDatePicker from '@/components/scr/scr-date-picker';
import MemberSelect from '@/components/MemberSelect';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { useRoute, useRouter } from 'vue-router';
import moment from 'moment';
import WName from '@/components/w-name';
import { Copy, DataShape, HelpDocumentFill } from 'bkui-vue/lib/icon';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import CommonSideslider from '@/components/common-sideslider';
import { timeFormatter, applicationTime } from '@/common/util';
import http from '@/http';
import { useZiyanScrStore, useUserStore } from '@/store';
import SuborderDetail from '../suborder-detail';
import CommonDialog from '@/components/common-dialog';
import { throttle } from 'lodash';
import MatchPanel from '../match-panel';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getResourceTypeName } from '../transform';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { FormItem } = Form;
export default defineComponent({
  setup() {
    const userStore = useUserStore();
    const businessMapStore = useBusinessMapStore();
    const { transformApplyStages } = useApplyStages();
    const { transformRequireTypes } = useRequireTypes();
    const isSidesliderShow = ref(false);
    const machineDetails = ref([]);
    const isMatchPanelShow = ref(false);
    const isDialogShow = ref(false);
    const curRow = ref({});
    const curSuborder = ref({
      step_name: '',
      step_id: 1,
      suborder_id: 0,
    });
    const scrStore = useZiyanScrStore();
    const { formModel, resetForm } = useFormModel({
      bkBizId: ['all'],
      requireType: [],
      stage: [],
      orderId: [],
      dateRange: applicationTime(),
      user: [userStore.username],
    });
    const reapply = (data: any) => {
      router.push({
        path: '/ziyanScr/hostApplication/apply',
        query: { order_id: data.order_id },
      });
    };
    const modify = (data: any) => {
      router.push({
        path: '/ziyanScr/hostApplication/modify',
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
        if (
          ['wait', 'MATCHED_SOME', 'MATCHING'].includes(row.status) ||
          (row.stage === 'SUSPEND' && row.status === 'TERMINATE')
        ) {
          return false;
        }
        if (['UNCOMMIT', 'PAUSED'].includes(row.status)) {
          return true;
        }
        if (['TERMINATE', 'AUDIT'].includes(row.stage) && !row.status) {
          return true;
        }
        if (row.stage === 'DONE' && row.status === 'DONE') {
          return true;
        }
        return false;
      };
    });
    const { CommonTable, getListData, isLoading } = useTable({
      tableOptions: {
        columns: [
          {
            label: '单号/子单号',
            width: 100,
            render: ({ data }: any) => {
              return (
                <div class={'flex-row align-item-center'}>
                  <Button
                    theme='primary'
                    text
                    onClick={() => {
                      router.push({
                        name: 'host-application-detail',
                        params: {
                          id: data.order_id,
                        },
                      });
                    }}>
                    {data.order_id}
                  </Button>
                  <br />
                  <p class={'ml8 sub-order-txt'}>子单号: {data.suborder_id || '无'}</p>
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
                  <Button size='small' onClick={() => modify(data)} text theme={'primary'}>
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
                      isSidesliderShow.value = true;
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
            render: ({ data }: any) => transformRequireTypes(data.requireType),
          },
          {
            label: '需求摘要',
            width: 250,
            render: ({ data }: any) => {
              return (
                <div>
                  <p>资源类型：{getResourceTypeName(data?.resource_type)}</p>
                  <p>机型：{data.spec?.device_type || '--'}</p>
                  <p>园区：{getZoneCn(data.spec?.zone)}</p>
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
          border: ['row', 'col', 'outer'],
          onRowMouseEnter: (e, row) => {
            handleCellMouseEnter(row);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => ({
        url: '/api/v1/woa/task/findmany/apply',
        payload: removeEmptyFields({
          bk_biz_id: formModel.bkBizId.length === 1 && formModel.bkBizId[0] === 'all' ? undefined : formModel.bkBizId,
          order_id: formModel.orderId.length
            ? String(formModel.orderId)
                .split('\n')
                .map((v) => +v)
            : undefined,
          // suborder_id: formModel.suborderId,
          bk_username: formModel.user,
          stage: formModel.stage,
          start: formModel.dateRange[0],
          end: formModel.dateRange[1],
          require_type: formModel.requireType,
        }),
      }),
    });

    // 获取匹配详情
    const getMatchDetails = async (subOrderId: number) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/apply/detail`, {
        suborder_id: subOrderId,
      });
    };
    // 已交付设备
    const getDeliveredDevices = (params) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/apply/device`, params);
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
        <div class={'filter-container'}>
          <Form model={formModel} class={'scr-form-wrapper'}>
            <FormItem label='业务'>
              <BusinessSelector autoSelect v-model={formModel.bkBizId} multiple authed isShowAll />
            </FormItem>
            <FormItem label='需求类型'>
              <RequirementTypeSelector v-model={formModel.requireType} multiple />
            </FormItem>
            <FormItem label='单据状态'>
              <ApplicationStatusSelector v-model={formModel.stage} multiple />
            </FormItem>
            <FormItem label='单号'>
              <Input
                v-model={formModel.orderId}
                type='textarea'
                autosize
                resize={false}
                placeholder='请输入单号,多个换行分割'
              />
            </FormItem>
            <FormItem label='申请时间'>
              <ScrDatePicker v-model={formModel.dateRange} />
            </FormItem>
            <FormItem label='申请人'>
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
            </FormItem>
          </Form>
        </div>
        <Button
          theme='primary'
          onClick={() => {
            router.push({
              path: '/ziyanScr/hostApplication/apply',
              query: route.query,
            });
          }}
          class={'ml24'}>
          新增申请
        </Button>
        <Button
          theme={'primary'}
          onClick={() => {
            getListData();
          }}
          class={'ml24 mr8'}
          loading={isLoading.value}>
          查询
        </Button>
        <Button
          onClick={() => {
            resetForm();
            getListData();
          }}>
          清空
        </Button>
        <div class={'table-container'}>
          <CommonTable />
        </div>

        <CommonSideslider v-model:isShow={isSidesliderShow.value} title='资源匹配详情' width={1000} noFooter>
          <Table
            showOverflowTooltip
            border={['outer', 'col', 'row']}
            data={machineDetails.value}
            columns={[
              {
                field: 'step_id',
                label: 'ID',
                width: '60',
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
                fixed: 'right',
              },
            ]}></Table>
        </CommonSideslider>

        <CommonDialog v-model:isShow={isDialogShow.value} title={`资源${curSuborder.value.step_name}详情`} width={800}>
          <SuborderDetail suborderId={curSuborder.value.suborder_id} stepId={curSuborder.value.step_id} />
        </CommonDialog>

        <CommonSideslider v-model:isShow={isMatchPanelShow.value} title='待匹配' width={1200} noFooter>
          <MatchPanel data={curRow.value} handleClose={() => (isMatchPanelShow.value = false)} />
        </CommonSideslider>
      </div>
    );
  },
});
