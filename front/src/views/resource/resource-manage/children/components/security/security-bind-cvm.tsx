import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { useAccountStore } from '@/store';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { DoublePlainObject } from '@/typings/resource';
import { Button, Dialog, Table, Loading, Message, Alert, InfoBox } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { defineComponent, onMounted, ref, watch } from 'vue';
import useSelection from '../../../hooks/use-selection';
import { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../../../common/table/HostOperations';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { useRegionsStore } from '@/store/useRegionsStore';
import useColumns from '../../../hooks/use-columns';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    detail: {
      type: Object,
      required: true,
    },
    sgId: {
      type: String,
      required: true,
    },
    sgCloudId: {
      type: String,
      required: true,
    }
  },
  setup(props) {
    const businessMapStore = useBusinessMapStore();
    const tableData = ref([]);
    const isDisassociateDisabled = ref(true);
    const isBindDialogLoading = ref(false);
    const isUnbindDialogLoading = ref(false);
    const regionsStore = useRegionsStore();
    const {generateColumnsSettings} = useColumns('cvms');
    const { whereAmI } = useWhereAmI();
    const tableColumns = [
      {
        field: 'selection',
        type: 'selection',
        width: '50',
      },
      {
        label: '内网IP',
        field: 'private_ipv4_addresses',
        render: ({ data }: any) => data.private_ipv4_addresses?.join(',')
          || data.private_ipv6_addresses?.join(',')
          || '--',
      },
      {
        label: '公网IP',
        field: 'public_ipv4_addresses',
        render: ({ data }: any) => data.public_ipv4_addresses?.join(',')
          || data.public_ipv6_addresses?.join(',')
          || '--',
      },
      {
        label: '云厂商',
        field: 'vendor',
      },
      {
        label: '地域',
        field: 'region',
        render: ({ cell, row }: any) => regionsStore.getRegionName(row.vendor, cell), 
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '状态',
        field: 'status',
        render({ data }: any) {
          return (
            <div class={'cvm-status-container'}>
              {HOST_SHUTDOWN_STATUS.includes(data.status) ? (
                <img
                  src={StatusAbnormal}
                  class={'mr6'}
                  width={13}
                  height={13}></img>
              ) : HOST_RUNNING_STATUS.includes(data.status) ? (
                <img
                  src={StatusNormal}
                  class={'mr6'}
                  width={13}
                  height={13}></img>
              ) : (
                <img
                  src={StatusUnknown}
                  class={'mr6'}
                  width={13}
                  height={13}></img>
              )}
              <span>{CLOUD_HOST_STATUS[data.status] || data.status}</span>
            </div>
          );
        },
      },
      {
        label: '是否分配',
        field: 'bk_biz_id',
        render: ({
          data,
          cell,
        }: {
          data: { bk_biz_id: number };
          cell: number;
        }) => (
          <bk-tag
            v-bk-tooltips={{
              content: businessMapStore.businessMap.get(cell),
              disabled: !cell || cell === -1,
            }}
            theme={data.bk_biz_id === -1 ? false : 'success'}>
            {data.bk_biz_id === -1 ? '未分配' : '已分配'}
          </bk-tag>
        ),
      },
      {
        label: '操作',
        field: 'operation',
        render: ({ data }: any) => (
          <Button text theme='primary' onClick={() => {
            InfoBox({
              title: '请确认是否解绑',
              subTitle: `将解绑【${data.private_ipv6_addresses?.join(',') || data.private_ipv4_addresses?.join(',')}】`,
              theme: 'danger',
              headerAlign: 'center',
              footerAlign: 'center',
              contentAlign: 'center',
              extCls: 'delete-resource-infobox',
              onConfirm: async () => {
                selections.value = [data];
                try {
                  await handleUnBind();
                }
                finally {
                  selections.value = [];
                }
              }
            });
          }} disabled={data.extension?.cloud_security_group_ids.length < 2} v-bk-tooltips={{
            content: '绑定的安全组少于2条,不能解绑',
            disabled: data.extension?.cloud_security_group_ids.length > 1,
          }}>
            解绑
          </Button>
        ),
      },
    ].filter(({field}) => (whereAmI.value === Senarios.business && field !== 'bk_biz_id') || whereAmI.value !== Senarios.business);
    const toBindCvmsListColumns = [
      {
        type: 'selection',
        width: '50',
        isDefaultShow: true,
      },
      {
        label: '主机ID',
        field: 'cloud_id',
      },
      {
        label: '内网IP',
        field: 'private_ipv4_addresses',
        render: ({ data }: any) => data.private_ipv4_addresses?.join(',')
          || data.private_ipv6_addresses?.join(',')
          || '--',
        isDefaultShow: true,
      },
      {
        label: '公网IP',
        field: 'public_ipv4_addresses',
        render: ({ data }: any) => data.public_ipv4_addresses?.join(',')
          || data.public_ipv6_addresses?.join(',')
          || '--',
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
                <img
                  src={StatusAbnormal}
                  class={'mr6'}
                  width={13}
                  height={13}></img>
              ) : HOST_RUNNING_STATUS.includes(data.status) ? (
                <img
                  src={StatusNormal}
                  class={'mr6'}
                  width={13}
                  height={13}></img>
              ) : (
                <img
                  src={StatusUnknown}
                  class={'mr6'}
                  width={13}
                  height={13}></img>
              )}
              <span>{CLOUD_HOST_STATUS[data.status] || data.status}</span>
            </div>
          );
        },
      },
      {
        label: '已绑定的安全组',
        field: 'cloud_security_group_ids',
        isDefaultShow: true,
        render: ({ data }: any) => (
          <div
            v-bk-tooltips={{ content: data.extension.security_group_names.join(',') }}
          >
            {
              data.extension.cloud_security_group_ids.join(',') || '--'
            }
          </div>
        ),
      },
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
    const {
      selections,
      handleSelectionChange,
    } = useSelection();

    const {
      selections: toBindCvmsListSelections,
      handleSelectionChange: handleToBindCvmsListSelectionChange,
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
      const url = whereAmI.value === Senarios.business
        ? `/api/v1/cloud/bizs/${accountStore.bizs}/cvms/security_groups/${props.sgId}`
        : `/api/v1/cloud/cvms/security_groups/${props.sgId}`;
      const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}${url}`);
      tableData.value = res?.data?.details || [];
      totalCount.value = res?.data?.count || 0;
      isTableLoading.value = false;
    };

    const getToBindCvms = async () => {
      if (![Senarios.business].includes(whereAmI.value)) return;
      isToBindCvmsTableLoading.value = true;
      const url = `/api/v1/cloud/bizs/${accountStore.bizs}/vendors/tcloud-ziyan/cmdb/hosts/list`;
      const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}${url}`, {
        account_id: props.detail.account_id,
        bk_biz_id: props.detail.bk_biz_id,
        query_from_cloud: true,
        page: {
          ...toBindCvmsListPagination.value,
          start: (toBindCvmsListPagination.value.current - 1) * toBindCvmsListPagination.value.limit,
        },
      });
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
        <BkButtonGroup class={'mb8'}>
          <Button selected={true}>云主机({totalCount.value})</Button>
        </BkButtonGroup>
        <div>
          {
            whereAmI.value === Senarios.business
              ? (
                <div class={'mb8'}>
                  <Button
                    theme='primary'
                    class={'mr8'}
                    onClick={() => (isBindDialogShow.value = true)}>
                    绑定
                  </Button>
                  <Button onClick={() => (isUnBindDialogShow.value = true)} disabled={isDisassociateDisabled.value}>
                    批量解绑
                  </Button>
                </div>
              ) : null
          }
          <Loading
            loading={isTableLoading.value}
            opacity={1}
            zIndex={99}
            color='#fff'>
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
              onSelectAll={(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)}
            ></Table>
          </Loading>
        </div>

        <Dialog
          isShow={isBindDialogShow.value}
          onClosed={() => (isBindDialogShow.value = false)}
          isLoading={isBindDialogLoading.value}
          width={1500}
          title='绑定主机'
        >
          {{
            default: () => (
              <Loading loading={isToBindCvmsTableLoading.value}>
                <Alert
                  theme='info'
                  class={'mb12'}
                  title='新绑定的安全组为最高优先级，如主机上已绑定的安全组名为“安全组1”，新绑定的安全组名为“安全组2”，则依次生效的安全组顺序为“安全组2”、“安全组1”'
                >
                </Alert>
                <Table
                  columns={toBindCvmsListColumns}
                  data={toBindCvmsList.value}
                  settings={toBindCvmsSetting.value}
                  remotePagination
                  isRowSelectEnable={isToBindCvmsRowSelectEnable}
                  onSelectionChange={(selections: any) => handleToBindCvmsListSelectionChange(selections, isToBindCvmsCurRowSelectEnable)}
                  onSelectAll={(selections: any) => handleToBindCvmsListSelectionChange(selections, isToBindCvmsRowSelectEnable, true)}
                  onPageLimitChange={handleToBindListPageLimitChange}
                  onPageValueChange={handleToBindListPageChange}
                  pagination={{
                    ...toBindCvmsListPagination.value,
                    count: toBindCvmsTotalCount.value,
                  }}></Table>
              </Loading>
            ),
            footer: () => (
              <div>
                <Button theme='primary' disabled={!toBindCvmsListSelections.value.length} onClick={handleBind}>确定</Button>
                <Button class={'ml8'} onClick={() => (isBindDialogShow.value = false)}>取消</Button>
              </div>
            )
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
            <Table columns={tableColumns.filter(({ field }) => !['selection', 'operation'].includes(field))} data={selections.value}></Table>
          </BkButtonGroup>
        </Dialog>
      </div>
    );
  },
});
