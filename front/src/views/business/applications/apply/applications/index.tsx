import { computed, defineComponent, ref, watch, onMounted } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button, Message, TagInput } from 'bkui-vue';
import { Copy, DataShape, HelpDocumentFill } from 'bkui-vue/lib/icon';
import GridFilterComp from '@/components/grid-filter-comp';
import ScrDatePicker from '@/components/scr/scr-date-picker';
import ScrCreateFilterSelector from '@/views/ziyanScr/resource-manage/create/ScrCreateFilterSelector';
import MemberSelect from '@/components/MemberSelect';
import WName from '@/components/w-name';
import StageDetailSideslider from './stage-detail';
import MatchSideslider from './match';

import moment from 'moment';
import { useI18n } from 'vue-i18n';
import { throttle } from 'lodash';
import { useUserStore, useZiyanScrStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useFormModel from '@/hooks/useFormModel';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import useSearchQs from '@/hooks/use-search-qs';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import { applicationTime } from '@/common/util';
import { getResourceTypeName } from '@/views/ziyanScr/hostApplication/components/transform';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import http from '@/http';
import { useSaveSearchRules } from '../../useSaveSearchRules';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const router = useRouter();
    const { t } = useI18n();
    const userStore = useUserStore();
    const scrStore = useZiyanScrStore();
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const route = useRoute();

    const { transformApplyStages } = useApplyStages();
    const { transformRequireTypes } = useRequireTypes();

    const stageDetailSidesliderRef = ref();
    const matchSidesliderRef = ref();

    const { formModel, resetForm } = useFormModel({
      bkBizId: [],
      requireType: [],
      stage: [],
      orderId: [],
      dateRange: applicationTime(),
      bkUsername: [userStore.username],
    });
    const curRow = ref({});

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
        query: route.query,
      };
      if (row.stage === 'UNCOMMIT') {
        routeParams = { name: 'applyCvm', query: { ...routeParams.query, order_id: row.order_id, unsubmitted: 1 } };
      }
      router.push(routeParams);
    };

    const modify = (data: any) => {
      router.push({ name: 'HostApplicationsModify', query: { ...data } });
    };

    const reapply = (data: any) => {
      router.push({ name: 'applyCvm', query: { order_id: data.order_id, unsubmitted: 0 } });
    };

    const throttleInfo = ref(null);
    // 已交付设备
    const getDeliveredDevices = (params: any) => {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
        params,
      );
    };
    // 查询交付IP和固号IP
    const getDeliveredHostField = (row: any, fieldKey: any) => {
      const params = {
        filter: {
          condition: 'AND',
          rules: [
            { field: 'suborder_id', operator: 'equal', value: row.suborder_id },
            { field: 'bk_biz_id', operator: 'in', value: [row.bk_biz_id] },
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
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/find/apply/detail`, {
        suborder_id: subOrderId,
      });
    };

    const { CommonTable, getListData, isLoading, pagination } = useTable({
      tableOptions: {
        columns: [
          {
            label: t('单号/子单号'),
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
            label: t('单据状态'),
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
                                {t('建议')}
                                <Button size='small' text theme={'primary'} class={'ml8'}>
                                  {t('修改需求重试')}
                                </Button>
                              </span>
                            ) : (
                              <span>
                                {t('请查看详情后联系')} <WName name={'BK助手'} class={'ml8'}></WName> {t('进行处理')}
                              </span>
                            )}
                          </span>
                        ),
                      }}>
                      {t('备货状态异常')} <HelpDocumentFill fill='#ffbb00' width={12} height={12} class={'ml4'} />
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
                      content: t('IDC物理机不支持修改,请联系ICR(IEG资源服务助手)'),
                      disabled: data.resource_type !== 'IDCPM',
                    }}
                    text
                    theme={'primary'}>
                    {t('修改需求重试')}
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
                    {t('查看详情')}
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
            label: t('需求类型'),
            field: 'require_type',
            width: 100,
            render: ({ cell }: any) => transformRequireTypes(cell),
          },
          {
            label: t('需求摘要'),
            width: 250,
            render: ({ data }: any) => {
              return (
                <div>
                  <div style={'height: 30px!important;line-height: 30px;'}>
                    {t('资源类型')}：{getResourceTypeName(data?.resource_type)}
                  </div>
                  <div style={'height: 20px!important;line-height: 20px;'}>
                    {t('机型')}：{data.spec?.device_type || '--'}
                  </div>
                  <div style={'height: 30px!important;line-height: 30px;'}>
                    {t('园区')}：{getZoneCn(data.spec?.zone)}
                  </div>
                </div>
              );
            },
          },
          {
            label: t('申请人'),
            render: ({ data }: any) => {
              return <WName name={data.bk_username}></WName>;
            },
          },
          {
            label: t('需求数'),
            field: 'total_num',
          },
          {
            label: t('待交付数'),
            field: 'pending_num',
            render({ cell, data }: any) {
              return cell ? (
                <Button
                  theme='primary'
                  text
                  onClick={() => {
                    curRow.value = data;
                    matchSidesliderRef.value.triggerShow(true);
                  }}>
                  {cell}
                </Button>
              ) : (
                cell
              );
            },
          },
          {
            label: t('已交付数'),
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
                      v-bk-tooltips={{ content: t('复制 IP') }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      class='mr8'
                      v-clipboard:copy={assetIds.join('\n')}
                      v-bk-tooltips={{ content: t('复制固资号') }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      onClick={() => goToCmdb(ips)}
                      v-bk-tooltips={{ content: t('去蓝鲸配置平台管理资源') }}>
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
            label: t('操作'),
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
                    {t('再次申请')}
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={opBtnDisabled.value(data)}
                    onClick={async () => {
                      await scrStore.retryOrder({ suborder_id: [data.suborder_id] });
                      Message({ theme: 'success', message: t('重试成功') });
                      getListData();
                    }}>
                    {t('重试')}
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={opBtnDisabled.value(data)}
                    onClick={async () => {
                      await scrStore.stopOrder({ suborder_id: [data.suborder_id] });
                      Message({ theme: 'success', message: t('终止成功') });
                      getListData();
                    }}>
                    {t('终止')}
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
        url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/apply`,
        payload: removeEmptyFields({
          bk_biz_id: [getBizsId()],
          order_id: formModel.orderId.map((v) => Number(v)),
          bk_username: formModel.bkUsername,
          stage: formModel.stage,
          start: formModel.dateRange[0],
          end: formModel.dateRange[1],
          require_type: formModel.requireType,
        }),
      }),
    });

    const searchRulesKey = 'host_apply_applications_rules';
    const searchQs = useSearchQs({
      key: 'initial_filter',
      properties: [
        { id: 'requireType', type: 'number', name: 'requireType', op: QueryRuleOPEnum.IN },
        { id: 'orderId', type: 'number', name: 'orderId', op: QueryRuleOPEnum.IN },
        { id: 'suborder_id', type: 'number', name: 'suborder_id', op: QueryRuleOPEnum.IN },
      ],
    });
    const filterOrders = (searchRulesStr?: string) => {
      // 合并默认条件值
      Object.assign(formModel, searchQs.get(route.query));
      // 回填
      if (searchRulesStr) {
        // 解决人员选择器搜索问题
        formModel.bkUsername.length > 0 &&
          userStore.setMemberDefaultList([...new Set([...userStore.memberDefaultList, ...formModel.bkUsername])]);
      }
      pagination.start = 0;
      getListData();
    };
    const { saveSearchRules, clearSearchRules } = useSaveSearchRules(searchRulesKey, filterOrders, formModel);

    const handleSearch = () => {
      // update query
      saveSearchRules();
    };

    const handleReset = () => {
      resetForm({ bkUsername: [userStore.username] });
      // update query
      clearSearchRules();
    };

    onMounted(() => {
      throttleDeliveredHostField();
    });

    watch(
      () => userStore.username,
      (username) => {
        if (route.query[searchRulesKey]) return;
        // 无搜索记录，设置申请人默认值
        formModel.bkUsername = [username];
      },
    );

    return () => (
      <>
        <GridFilterComp
          onSearch={handleSearch}
          onReset={handleReset}
          loading={isLoading.value}
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
                  placeholder={t('请输入单号')}
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
                  v-model={formModel.bkUsername}
                  clearable
                  defaultUserlist={userStore.memberDefaultList.map((username) => ({
                    username,
                    display_name: username,
                  }))}
                  placeholder={t('请输入企业微信名')}
                />
              ),
            },
          ]}
        />
        <section class={cssModule['table-wrapper']}>
          <CommonTable />
        </section>
        <StageDetailSideslider ref={stageDetailSidesliderRef} details={machineDetails.value} />
        <MatchSideslider ref={matchSidesliderRef} data={curRow.value} />
      </>
    );
  },
});
