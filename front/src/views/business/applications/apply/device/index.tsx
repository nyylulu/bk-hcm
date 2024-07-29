import { computed, defineComponent } from 'vue';
import cssModule from './index.module.scss';

import { Button, DatePicker, Input, TagInput } from 'bkui-vue';
import ScrCreateFilterSelector from '@/views/ziyanScr/resource-manage/create/ScrCreateFilterSelector';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import WName from '@/components/w-name';
import GridFilterComp from '@/components/grid-filter-comp';

import { useI18n } from 'vue-i18n';
import { useUserStore, useZiyanScrStore } from '@/store';
import useFormModel from '@/hooks/useFormModel';
import { useTable } from '@/hooks/useTable/useTable';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';
import { applicationTime, timeFormatter } from '@/common/util';
import { useRouter } from 'vue-router';
import { getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export default defineComponent({
  setup() {
    const router = useRouter();
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const { t } = useI18n();
    const userStore = useUserStore();
    const scrStore = useZiyanScrStore();

    const { formModel, resetForm } = useFormModel({
      orderId: '',
      bkBizId: [],
      bkUsername: [userStore.username],
      ip: [],
      requireType: '',
      suborderId: '',
      dateRange: applicationTime(),
    });

    const { selections, handleSelectionChange } = useSelection();

    const clipHostIp = computed(() => {
      return selections.value.map((item) => item.ip).join('\n');
    });
    const clipHostAssetId = computed(() => {
      return selections.value.map((item) => item.asset_id).join('\n');
    });

    const columns = [
      { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
      {
        label: '单号',
        field: 'order_id',
        width: 80,
        render: ({ data, cell }: any) => {
          return (
            <Button
              text
              theme='primary'
              onClick={() => {
                router.push({
                  name: 'HostApplicationsDetail',
                  params: { id: data.order_id },
                });
              }}>
              {cell}
            </Button>
          );
        },
      },
      { label: '子单号', field: 'suborder_id', width: 80 },
      { label: '需求类型', field: 'require_type', render: ({ row }: any) => getTypeCn(row.require_type) },
      {
        label: '申请人',
        field: 'bk_username',
        render({ cell }: any) {
          return <WName name={cell} />;
        },
      },
      { label: '内网IP', field: 'ip' },
      { label: '固资号', field: 'asset_id' },
      { label: '资源类型', field: 'resource_type' },
      { label: '机型', field: 'device_type' },
      { label: '园区', field: 'zone_name' },
      { label: '交付时间', field: 'update_at', render: ({ cell }: any) => timeFormatter(cell) },
      { label: '申请时间', field: 'create_at', render: ({ cell }: any) => timeFormatter(cell) },
      {
        label: '备注信息',
        field: 'remark',
        render({ data }: any) {
          return `${data.description}${data.description && data.remark && '/'}${data.remark}` || '--';
        },
      },
    ];

    const { CommonTable, getListData, isLoading, dataList, pagination } = useTable({
      tableOptions: {
        columns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
          payload: {
            filter: transferSimpleConditions([
              'AND',
              ['bk_biz_id', 'in', [getBizsId()]],
              ['require_type', '=', formModel.requireType],
              ['order_id', '=', formModel.orderId],
              ['suborder_id', '=', formModel.suborderId],
              ['bk_username', 'in', formModel.bkUsername],
              ['ip', 'in', formModel.ip],
              ['update_at', 'd>=', timeFormatter(formModel.dateRange[0], 'YYYY-MM-DD')],
              ['update_at', 'd<=', timeFormatter(formModel.dateRange[1], 'YYYY-MM-DD')],
            ]),
            page: { start: 0, limit: 10 },
          },
        };
      },
    });

    const filterOrders = () => {
      pagination.start = 0;
      formModel.bkBizId = [getBizsId()];
      getListData();
    };

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
                  multiple={false}
                  optionIdPath='require_type'
                  optionNamePath='require_name'
                />
              ),
            },
            {
              title: t('单号'),
              content: <Input v-model={formModel.orderId} clearable type='number' placeholder='请输入单号' />,
            },
            {
              title: t('申请人'),
              content: (
                <MemberSelect
                  v-model={formModel.bkUsername}
                  multiple
                  clearable
                  defaultUserlist={[{ username: userStore.username, display_name: userStore.username }]}
                  placeholder={t('请输入企业微信名')}
                />
              ),
            },
            {
              title: t('交付时间'),
              content: <DatePicker class='full-width' type='daterange' v-model={formModel.dateRange} />,
            },
            {
              title: t('内网IP'),
              content: (
                <TagInput
                  v-model={formModel.ip}
                  allow-create
                  collapse-tags
                  allow-auto-match
                  pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
                  createTagValidator={(ip) =>
                    /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/.test(ip)
                  }
                  placeholder='输入合法的 IP 地址'
                />
              ),
            },
          ]}
          onSearch={filterOrders}
          onReset={() => {
            resetForm({ user: [userStore.username] });
            formModel.bkBizId = [getBizsId()];
            filterOrders();
          }}
          loading={isLoading.value}
          immediate
        />
        <section class={cssModule['table-wrapper']}>
          <div class={[cssModule.buttons, cssModule.mb16]}>
            <ExportToExcelButton
              class={cssModule.button}
              data={selections.value}
              columns={columns}
              filename='设备列表'
            />
            <ExportToExcelButton
              class={cssModule.button}
              data={dataList.value}
              columns={columns}
              filename='设备列表'
              text='导出全部'
            />
            <Button class={cssModule.button} v-clipboard={clipHostIp.value} disabled={selections.value.length === 0}>
              复制IP
            </Button>
            <Button
              class={cssModule.button}
              v-clipboard={clipHostAssetId.value}
              disabled={selections.value.length === 0}>
              复制固单号
            </Button>
          </div>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </>
    );
  },
});
