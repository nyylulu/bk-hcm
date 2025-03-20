import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { useAccountStore, useResourceStore } from '@/store';
import { DoublePlainObject } from '@/typings/resource';
import { Button, Dialog, Table, Loading, Message, Alert, Input, ResizeLayout, Exception } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { defineComponent, onMounted, PropType, ref, watch } from 'vue';
import useSelection from '../../../../hooks/use-selection';
import { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../../../../common/table/HostOperations';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import useColumns from '../../../../hooks/use-columns';
import { useI18n } from 'vue-i18n';
import './index.scss';
import { useRoute } from 'vue-router';
import { QueryRuleOPEnum } from '@/typings';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    detail: {
      type: Object as PropType<{
        [key: string]: any;
        account_id: string;
        region: string;
        vendor: string;
      }>,
      required: true,
    },
    sgId: {
      type: String,
      required: true,
    },
    sgCloudId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const tableData = ref([]);
    const isDisassociateDisabled = ref(true);
    const isBindDialogLoading = ref(false);
    const isUnbindDialogLoading = ref(false);
    const { generateColumnsSettings } = useColumns('cvms');
    const { whereAmI } = useWhereAmI();
    const resourceStore = useResourceStore();
    const { t } = useI18n();
    const toBindCvmsListRef = ref();
    const { columns: tableColumns } = useColumns('cvms');
    const toBindCvmsListColumns = [
      { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
      {
        label: '主机ID',
        field: 'cloud_id',
      },
      {
        label: '内网IP',
        field: 'private_ipv4_addresses',
        render: ({ data }: any) =>
          data.private_ipv4_addresses?.join(',') || data.private_ipv6_addresses?.join(',') || '--',
        isDefaultShow: true,
      },
      {
        label: '公网IP',
        field: 'public_ipv4_addresses',
        render: ({ data }: any) =>
          data.public_ipv4_addresses?.join(',') || data.public_ipv6_addresses?.join(',') || '--',
        isDefaultShow: true,
      },
      {
        label: '名称',
        field: 'name',
        isDefaultShow: true,
      },
      {
        label: '所属网络',
        field: 'cloud_vpc_ids',
        render: ({ data }: any) => data.cloud_vpc_ids.join(',') || '--',
        isDefaultShow: true,
      },
      {
        label: '状态',
        field: 'status',
        isDefaultShow: true,
        render({ data }: any) {
          return (
            <div class={'cvm-status-container'}>
              {HOST_SHUTDOWN_STATUS.includes(data.status) ? (
                <img src={StatusAbnormal} class={'mr6'} width={13} height={13}></img>
              ) : HOST_RUNNING_STATUS.includes(data.status) ? (
                <img src={StatusNormal} class={'mr6'} width={13} height={13}></img>
              ) : (
                <img src={StatusUnknown} class={'mr6'} width={13} height={13}></img>
              )}
              <span>{CLOUD_HOST_STATUS[data.status] || data.status}</span>
            </div>
          );
        },
      },
      // {
      //   label: '已绑定的安全组',
      //   field: 'cloud_security_group_ids',
      //   isDefaultShow: true,
      //   render: ({ data }: any) => (
      //     <div v-bk-tooltips={{ content: data.extension.security_group_names.join(',') }}>
      //       {data.extension.cloud_security_group_ids.join(',') || '--'}
      //     </div>
      //   ),
      // },
    ];
    const toBindCvmsSetting = generateColumnsSettings(toBindCvmsListColumns);
    const totalCount = ref(0);
    const pagination = ref({
      start: 0,
      limit: 10,
      current: 1,
    });
    const toBindCvmsListPagination = ref({
      start: 0,
      limit: 10,
      current: 1,
    });
    const toBindCvmsTotalCount = ref(0);
    const toBindCvmsList = ref([]);
    const isBindDialogShow = ref(false);
    const isUnBindDialogShow = ref(false);
    const isTableLoading = ref(false);
    const isToBindCvmsTableLoading = ref(false);
    const accountStore = useAccountStore();
    const { selections, handleSelectionChange } = useSelection();
    const route = useRoute();
    const { getBusinessApiPath } = useWhereAmI();

    const {
      selections: toBindCvmsListSelections,
      handleSelectionChange: handleToBindCvmsListSelectionChange,
      resetSelections: resetToBindCvmsSelections,
    } = useSelection();

    const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => {
      return row?.extension?.cloud_security_group_ids.length > 1;
    };

    const isToBindCvmsRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isToBindCvmsCurRowSelectEnable(row);
    };
    const isToBindCvmsCurRowSelectEnable = (row: any) => {
      return !row?.extension?.cloud_security_group_ids?.includes(props.sgCloudId);
    };

    const getTableList = async () => {
      isTableLoading.value = true;
      try {
        const [detailsRes, countRes] = await Promise.all(
          [false, true].map((isCount) =>
            http.post(
              `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}security_group/${route.query.id}/cvm/list`,
              {
                filter: {
                  op: QueryRuleOPEnum.AND,
                  rules: [{ field: 'security_group_id', op: QueryRuleOPEnum.EQ, value: route.query.id }],
                },
                page: {
                  start: isCount ? 0 : pagination.value.start,
                  limit: isCount ? 0 : pagination.value.limit,
                  sort: isCount ? undefined : 'created_at',
                  order: isCount ? undefined : 'DESC',
                  count: isCount,
                },
              },
            ),
          ),
        );
        const cvmIds = detailsRes.data?.details?.map((cvm: any) => cvm.cvm_id) ?? [];
        if (cvmIds.length > 0) {
          const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}cvms/list`, {
            filter: {
              op: QueryRuleOPEnum.AND,
              rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: cvmIds }],
            },
            page: {
              start: 0,
              limit: pagination.value.limit,
              count: false,
            },
          });
          tableData.value = res.data.details;
          totalCount.value = countRes.data.count;
        }
      } catch (error) {
        console.error(error);
        tableData.value = [];
        totalCount.value = 0;
      } finally {
        isTableLoading.value = false;
      }
    };

    const getToBindCvms = async () => {
      if (![Senarios.business].includes(whereAmI.value)) return;
      isToBindCvmsTableLoading.value = true;
      const res = await resourceStore.list(
        {
          filter: {
            op: 'and',
            rules: [
              {
                field: 'account_id',
                op: 'eq',
                value: props.detail.account_id,
              },
              {
                field: 'region',
                op: 'eq',
                value: props.detail.region,
              },
              {
                field: 'vendor',
                op: 'eq',
                value: props.detail.vendor,
              },
            ],
          },
          page: {
            count: false,
            limit: pagination.value.limit,
            sort: 'created_at',
            order: 'DESC',
          },
        },
        'cvms',
      );

      toBindCvmsList.value = res.data?.details;
      toBindCvmsTotalCount.value = res.data?.count;
      isToBindCvmsTableLoading.value = false;
    };

    const handleBind = async () => {
      isBindDialogLoading.value = true;
      const url = `/api/v1/cloud/bizs/${accountStore.bizs}/security_groups/associate/cloud_cvms/batch`;
      await http.post(`${BK_HCM_AJAX_URL_PREFIX}${url}`, {
        bk_biz_id: accountStore.bizs,
        security_group_id: props.sgId,
        cloud_cvm_ids: toBindCvmsListSelections.value.map(({ cloud_id }) => cloud_id),
      });
      Message({
        theme: 'success',
        message: '绑定成功',
      });
      isBindDialogLoading.value = false;
      isBindDialogShow.value = false;
      getTableList();
      getToBindCvms();
    };

    const handleUnBind = async () => {
      isUnbindDialogLoading.value = true;
      const url = `/api/v1/cloud/bizs/${accountStore.bizs}/security_groups/disassociate/cloud_cvms/batch`;
      await http.post(`${BK_HCM_AJAX_URL_PREFIX}${url}`, {
        bk_biz_id: accountStore.bizs,
        security_group_id: props.sgId,
        cloud_cvm_ids: selections.value.map(({ cloud_id }) => cloud_id),
      });
      Message({
        theme: 'success',
        message: '解除绑定成功',
      });
      isUnBindDialogShow.value = false;
      isUnbindDialogLoading.value = false;
      getTableList();
      getToBindCvms();
    };

    onMounted(() => {
      getTableList();
      getToBindCvms();
    });

    watch(
      () => selections.value,
      (val) => {
        isDisassociateDisabled.value = !val.length;
      },
      {
        deep: true,
      },
    );

    const handlePageChange = (current: number) => {
      pagination.value.current = current;
      getTableList();
    };

    const handlePageLimitChange = (limit: number) => {
      pagination.value.limit = limit;
      getTableList();
    };

    const handleToBindListPageChange = (current: number) => {
      toBindCvmsListPagination.value.current = current;
      getToBindCvms();
    };

    const handleToBindListPageLimitChange = (limit: number) => {
      toBindCvmsListPagination.value.limit = limit;
      getToBindCvms();
    };

    return () => (
      <div>
        <div>
          {whereAmI.value === Senarios.business ? (
            <div class={'mb8'}>
              <Button theme='primary' class={'mr8'} onClick={() => (isBindDialogShow.value = true)}>
                绑定
              </Button>
              <Button onClick={() => (isUnBindDialogShow.value = true)} disabled={isDisassociateDisabled.value}>
                批量解绑
              </Button>
            </div>
          ) : null}
          <Loading loading={isTableLoading.value} opacity={1} zIndex={99} color='#fff'>
            <Table
              columns={tableColumns}
              data={tableData.value}
              remotePagination
              pagination={{
                ...pagination.value,
                count: totalCount.value,
              }}
              onPageLimitChange={handlePageLimitChange}
              onPageValueChange={handlePageChange}
              isRowSelectEnable={isRowSelectEnable}
              onSelectionChange={(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)}
              onSelectAll={(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)}></Table>
          </Loading>
        </div>

        <Dialog
          isShow={isBindDialogShow.value}
          onClosed={() => (isBindDialogShow.value = false)}
          isLoading={isBindDialogLoading.value}
          class={'bind-cvm-dialog-container'}
          width={1500}
          title='绑定主机'>
          {{
            default: () => (
              <Loading loading={isToBindCvmsTableLoading.value}>
                <ResizeLayout placement='right' initialDivide={'25%'} class={'security-bind-cvm-wrapper'}>
                  {{
                    aside: () => (
                      <div class={'aside-container'}>
                        <div class={'aside-header'}>{t('结果预览')}</div>
                        <div class={'aside-content'}>
                          <div class={'aside-content-toolbar'}>
                            {t('已选择主机')}
                            <div class={'total-sum'}>{toBindCvmsListSelections.value.length}</div>
                            <div
                              class={'clear-tool'}
                              onClick={() => {
                                resetToBindCvmsSelections();
                                toBindCvmsListRef.value.clearSelection();
                              }}>
                              <i class={'hcm-icon bkhcm-icon-cc-clear mr4'}></i>
                              {t('清空')}
                            </div>
                          </div>
                          <div class={'aside-content-list'}>
                            {toBindCvmsListSelections.value.length ? (
                              toBindCvmsListSelections.value.map(({ private_ipv4_addresses }, idx) => (
                                <div class={'list-item'}>
                                  {private_ipv4_addresses}
                                  <i
                                    class={'hcm-icon bkhcm-icon-close close-icon'}
                                    onClick={() => {
                                      toBindCvmsListSelections.value.splice(idx, 1);
                                      toBindCvmsListRef.value.toggleRowSelection(
                                        toBindCvmsListRef.value.getSelection?.()?.[idx],
                                        false,
                                      );
                                    }}
                                  />
                                </div>
                              ))
                            ) : (
                              <Exception description='暂未选中' scene='part' type='empty' />
                            )}
                          </div>
                        </div>
                      </div>
                    ),
                    main: () => (
                      <div class={'bind-cvm-container'}>
                        <Alert
                          theme='warning'
                          class={'mb16'}
                          title='新绑定的安全组为最高优先级，如主机上已绑定的安全组名为「安全组」，新绑定的安全组名为「安全组2」，则依次生效的安全组顺序为：安全组2、安全组1'></Alert>
                        <Input placeholder={t('请输入IP')} class={'mb16'} />
                        <Table
                          ref={toBindCvmsListRef}
                          columns={toBindCvmsListColumns}
                          data={toBindCvmsList.value}
                          settings={toBindCvmsSetting.value}
                          remotePagination
                          isRowSelectEnable={isToBindCvmsRowSelectEnable}
                          onSelectionChange={(selections: any) =>
                            handleToBindCvmsListSelectionChange(selections, isToBindCvmsCurRowSelectEnable)
                          }
                          onSelectAll={(selections: any) =>
                            handleToBindCvmsListSelectionChange(selections, isToBindCvmsRowSelectEnable, true)
                          }
                          onPageLimitChange={handleToBindListPageLimitChange}
                          onPageValueChange={handleToBindListPageChange}
                          pagination={{
                            ...toBindCvmsListPagination.value,
                            count: toBindCvmsTotalCount.value,
                          }}
                        />
                      </div>
                    ),
                  }}
                </ResizeLayout>
              </Loading>
            ),
            footer: () => (
              <div>
                <Button theme='primary' disabled={!toBindCvmsListSelections.value.length} onClick={handleBind}>
                  确定
                </Button>
                <Button class={'ml8'} onClick={() => (isBindDialogShow.value = false)}>
                  取消
                </Button>
              </div>
            ),
          }}
        </Dialog>

        <Dialog
          isShow={isUnBindDialogShow.value}
          title='批量解绑'
          width={1500}
          isLoading={isUnbindDialogLoading.value}
          onClosed={() => (isUnBindDialogShow.value = false)}
          onConfirm={handleUnBind}>
          <BkButtonGroup>
            <Table
              columns={tableColumns.filter(({ field }) => !['selection', 'operation'].includes(field))}
              data={selections.value}></Table>
          </BkButtonGroup>
        </Dialog>
      </div>
    );
  },
});
