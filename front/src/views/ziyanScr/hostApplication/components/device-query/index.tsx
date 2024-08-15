import { defineComponent, computed, watch, ref } from 'vue';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import BusinessSelector from '@/components/business-selector/index.vue';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';
import { Button, Form, Input } from 'bkui-vue';
import MemberSelect from '@/components/MemberSelect';
import useFormModel from '@/hooks/useFormModel';
import { useUserStore } from '@/store';
import { timeFormatter, applicationTime } from '@/common/util';
import ExportToExcelButton from '@/components/export-to-excel-button';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
const { FormItem } = Form;
export default defineComponent({
  setup() {
    const { columns } = useColumns('DeviceQuerycolumns');
    const { selections, handleSelectionChange } = useSelection();
    const clipHostIp = computed(() => {
      return selections.value.map((item) => item.ip).join('\n');
    });
    const clipHostAssetId = computed(() => {
      return selections.value.map((item) => item.asset_id).join('\n');
    });
    const userStore = useUserStore();
    const { formModel, resetForm } = useFormModel({
      orderId: '',
      bkBizId: [],
      bkUsername: [userStore.username],
      ip: '',
      requireType: '',
      suborderId: '',
      dateRange: applicationTime(),
    });

    const businessSelectorRef = ref();
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
          url: '/api/v1/woa/task/findmany/apply/device',
          payload: {
            filter: transferSimpleConditions([
              'AND',
              [
                'bk_biz_id',
                'in',
                formModel.bkBizId.length === 0
                  ? businessSelectorRef.value.businessList.slice(1).map((item: any) => item.id)
                  : formModel.bkBizId,
              ],
              ['require_type', '=', formModel.requireType],
              ['order_id', '=', formModel.orderId],
              ['suborder_id', '=', formModel.suborderId],
              ['bk_username', 'in', formModel.bkUsername],
              ['ip', 'in', ipArray.value],
              ['update_at', 'd>=', timeFormatter(formModel.dateRange[0], 'YYYY-MM-DD')],
              ['update_at', 'd<=', timeFormatter(formModel.dateRange[1], 'YYYY-MM-DD')],
            ]),
            page: { start: 0, limit: 10 },
          },
        };
      },
    });

    const ipArray = computed(() => {
      const ipv4 = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
      const ips = [];
      formModel.ip
        .split(/\r?\n/)
        .map((ip) => ip.trim())
        .filter((ip) => ip.length > 0)
        .forEach((item) => {
          if (ipv4.test(item)) {
            ips.push(item);
          }
        });
      return ips;
    });

    const filterOrders = () => {
      pagination.start = 0;
      formModel.bkBizId = formModel.bkBizId.length === 1 && formModel.bkBizId[0] === 'all' ? [] : formModel.bkBizId;
      getListData();
    };

    watch(
      () => businessSelectorRef.value?.businessList,
      (val) => {
        if (!val?.length) return;
        getListData();
      },
      { deep: true },
    );

    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'}>
          <Form label-width='110' formType='vertical' class='scr-form-wrapper' model={formModel}>
            <FormItem label='业务'>
              <BusinessSelector
                ref={businessSelectorRef}
                autoSelect
                multiple
                v-model={formModel.bkBizId}
                authed
                isShowAll
                notAutoSelectAll
                saveBizs
                bizsKey='scr_host_bizs'
              />
            </FormItem>
            <FormItem label='需求类型'>
              <RequirementTypeSelector v-model={formModel.requireType} />
            </FormItem>
            <FormItem label='单号'>
              <bk-input v-model={formModel.orderId} clearable type='number' placeholder='请输入单号'></bk-input>
            </FormItem>
            <FormItem label='申请人'>
              <MemberSelect
                v-model={formModel.bkUsername}
                multiple
                clearable
                defaultUserlist={[
                  {
                    username: userStore.username,
                    display_name: userStore.username,
                  },
                ]}
              />
            </FormItem>
            <FormItem label='交付时间'>
              <bk-date-picker type='daterange' v-model={formModel.dateRange} />
            </FormItem>
            <FormItem label='内网 IP'>
              <Input
                class={'filte-item'}
                type='textarea'
                clearable
                placeholder='请输入IP地址，多行换行分割'
                v-model={formModel.ip}
                autosize
                resize={false}
              />
            </FormItem>
          </Form>
          <div class='btn-container'>
            <Button theme='primary' native-type='submit' loading={isLoading.value} onClick={filterOrders}>
              查询
            </Button>
            <Button
              onClick={() => {
                resetForm();
                // 因为要保存业务全选的情况, 所以这里 defaultBusiness 可能是 ['all'], 而组件的全选对应着 [], 所以需要额外处理
                // 根源是此处的接口要求全选时携带传递所有业务id, 所以需要与空数组做区分
                formModel.bkBizId = businessSelectorRef.value.defaultBusiness;
                filterOrders();
              }}>
              重置
            </Button>
          </div>
        </div>
        <div class='btn-container oper-btn-pad'>
          <ExportToExcelButton data={selections.value} columns={columns} filename='设备列表' />
          <ExportToExcelButton data={dataList.value} columns={columns} filename='设备列表' text='导出全部' />
          <Button v-clipboard={clipHostIp.value} disabled={selections.value.length === 0}>
            复制IP
          </Button>
          <Button v-clipboard={clipHostAssetId.value} disabled={selections.value.length === 0}>
            复制固单号
          </Button>
        </div>
        <CommonTable />
      </div>
    );
  },
});
