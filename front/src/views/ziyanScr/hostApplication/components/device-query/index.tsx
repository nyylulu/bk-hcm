import { defineComponent, computed } from 'vue';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import BusinessSelector from '@/components/business-selector/index.vue';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';
import { Button, Form, Input } from 'bkui-vue';
import MemberSelect from '@/components/MemberSelect';
import useFormModel from '@/hooks/useFormModel';
import { timeFormatter } from '@/common/util';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import ExportToExcelButton from '@/components/export-to-excel-button';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
const { FormItem } = Form;
export default defineComponent({
  setup() {
    const { columns } = useColumns('DeviceQuerycolumns');
    const businessMapStore = useBusinessMapStore();
    const { selections, handleSelectionChange } = useSelection();
    const clipHostIp = computed(() => {
      return selections.value.map((item) => item.ip).join('\n');
    });
    const { formModel, resetForm } = useFormModel({
      orderId: '',
      bkBizId: businessMapStore.authedBusinessList?.[0]?.id,
      bkUsername: [],
      ip: '',
      requireType: '',
      suborderId: '',
      dateRange: [],
    });

    const { CommonTable, getListData, isLoading, dataList } = useTable({
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
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/task/findmany/apply/device',
          payload: {
            filter: transferSimpleConditions([
              'AND',
              ['bk_biz_id', '=', formModel.bkBizId],
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
    return () => (
      <div class={'device-query-container'}>
        <CommonTable>
          {{
            tabselect: () => (
              <>
                <Form label-width='110' class='scr-form-wrapper' model={formModel}>
                  <FormItem label='业务'>
                    <BusinessSelector autoSelect v-model={formModel.bkBizId} authed />
                  </FormItem>
                  <FormItem label='需求类型'>
                    <RequirementTypeSelector v-model={formModel.requireType} />
                  </FormItem>
                  <FormItem label='单号'>
                    <bk-input v-model={formModel.orderId} clearable type='number' placeholder='请输入单号'></bk-input>
                  </FormItem>
                  <FormItem label='申请人'>
                    <MemberSelect v-model={formModel.bkUsername} multiple clearable />
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

                  <Button
                    class={'ml24 mr8'}
                    theme='primary'
                    native-type='submit'
                    loading={isLoading.value}
                    onClick={() => {
                      getListData();
                    }}>
                    查询
                  </Button>
                  <Button
                    class={'mr8'}
                    onClick={() => {
                      resetForm();
                      getListData();
                    }}>
                    清空
                  </Button>

                  <ExportToExcelButton class={'mr8'} data={selections.value} columns={columns} filename='设备列表' />
                  <ExportToExcelButton
                    class={'mr8'}
                    data={dataList.value}
                    columns={columns}
                    filename='设备列表'
                    text='导出全部'
                  />
                  {/* <Button class={'mr8'}>导出全部</Button> */}
                  <Button class={'mr8'} v-clipboard={clipHostIp.value} disabled={selections.value.length === 0}>
                    复制
                  </Button>
                </Form>
              </>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
