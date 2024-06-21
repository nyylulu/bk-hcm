import { defineComponent, computed } from 'vue';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import BusinessSelector from '@/components/business-selector/index.vue';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';
import { Button, Form, Input } from 'bkui-vue';
import MemberSelect from '@/components/MemberSelect';
import useFormModel from '@/hooks/useFormModel';
import { useUserStore } from '@/store';
import { timeFormatter, applicationTime } from '@/common/util';
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
    const clipHostAssetId = computed(() => {
      return selections.value.map((item) => item.asset_id).join('\n');
    });
    const userStore = useUserStore();
    const { formModel, resetForm } = useFormModel({
      orderId: '',
      bkBizId: businessMapStore.authedBusinessList?.[0]?.id,
      bkUsername: [userStore.username],
      ip: '',
      requireType: '',
      suborderId: '',
      dateRange: applicationTime(),
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
              ['bk_biz_id', '=', formModel.bkBizId === 'all' ? '' : formModel.bkBizId],
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
                    复制IP
                  </Button>
                  <Button class={'mr8'} v-clipboard={clipHostAssetId.value} disabled={selections.value.length === 0}>
                    复制固单号
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
