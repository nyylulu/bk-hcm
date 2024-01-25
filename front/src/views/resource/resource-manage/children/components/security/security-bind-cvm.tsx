import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { useAccountStore } from '@/store';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { DoublePlainObject } from '@/typings/resource';
import { Button, Dialog, Table, Loading, Message } from 'bkui-vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { defineComponent, onMounted, ref, watch } from 'vue';
import useSelection from '../../../hooks/use-selection';
import { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../../../common/table/HostOperations';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';

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
  },
  setup(props) {
    const businessMapStore = useBusinessMapStore();
    const tableData = ref([]);
    const isDisassociateDisabled = ref(true);
    const isBindDialogLoading = ref(false);
    const isUnbindDialogLoading = ref(false);
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
          <Button text theme='primary' onClick={() => console.log(data)} disabled={data.extension?.cloud_security_group_ids.length < 2} v-bk-tooltips={{
            content: '绑定的安全组少于2条,不能解绑',
            disabled: data.extension?.cloud_security_group_ids.length > 1,
          }}>
            解绑
          </Button>
        ),
      },
    ];
    const toBindCvmsListColumns = [
      {
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
        label: '名称',
        field: 'name',
      },
      {
        label: '所属网络',
        field: 'associatedNetwork',
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
        label: '已绑定的安全组',
        field: 'boundSecurityGroup',
      },
    ];
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
    const { whereAmI } = useWhereAmI();
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
      return row.extension?.cloud_security_group_ids.length > 1;
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
      const url = `/api/v1/cloud/vendors/tcloud-ziyan/bizs/${accountStore.bizs}/cmdb/hosts/list`;
      const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}${url}`, {
        account_id: props.detail.account_id,
        bk_biz_id: props.detail.bk_biz_id,
        page: toBindCvmsListPagination.value,
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
          onConfirm={handleBind}
          width={1500}
          title='绑定主机'
        >
          <Loading loading={isToBindCvmsTableLoading.value}>
            <Table
              columns={toBindCvmsListColumns}
              data={toBindCvmsList.value}
              remotePagination
              onSelectionChange={(selections: any) => handleToBindCvmsListSelectionChange(selections, () => true)}
              onSelectAll={(selections: any) => handleToBindCvmsListSelectionChange(selections, () => true, true)}
              pagination={{
                ...toBindCvmsListPagination.value,
                count: toBindCvmsTotalCount.value,
              }}></Table>
          </Loading>
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
