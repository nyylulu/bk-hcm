import { defineComponent, ref, watch } from 'vue';
import ResourceSelect from './components/ResourceSelect';
import ResourceType from './components/ResourceType';
import ResourceConfirm from './components/ResourceConfirm';
import { Dialog, Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '@/components/common-sideslider';
import './index.scss';
export default defineComponent({
  name: 'RecyclingResources',
  setup() {
    const active = ref(2);
    const objectSteps = ref([{ title: '输入IP/固资' }, { title: '确认回收类型' }, { title: '信息确认与提交' }]);
    const tableHosts = ref([]);
    const bkBizId = ref();
    const tableSelectedHosts = ref([]);
    const recycleForm = ref({
      ips: [],
      remark: '',
    });
    const dialogVisible = ref(false);
    const selectedHosts = ref([]);
    const returnPlan = ref({
      cvm: '',
      pm: '',
      skipConfirm: '',
    });
    const drawer = ref(false);
    const ResourcesTotal = ref(false);
    const lips = ref();
    const activetab = ref(0);
    const { columns: BScolumns } = useColumns('BusinessSelection');
    const { columns: RTcolumns } = useColumns('ResourcesTotal');
    const { CommonTable: BSCommonTable, getListData } = useTable({
      tableOptions: {
        columns: BScolumns,
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: true,
          interface: {
            Parameters: {
              filter: {
                condition: 'AND',
                rules: [
                  {
                    field: 'require_type',
                    operator: 'equal',
                    value: 1,
                  },
                  {
                    field: 'label.device_group',
                    operator: 'in',
                    value: ['标准型'],
                  },
                ],
              },
              page: { limit: 0, count: 0 },
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const { CommonTable: RTCommonTable } = useTable({
      tableOptions: {
        columns: RTcolumns,
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: false,
          interface: {
            Parameters: {
              filter: {
                condition: 'AND',
                rules: [
                  {
                    field: 'require_type',
                    operator: 'equal',
                    value: 1,
                  },
                  {
                    field: 'label.device_group',
                    operator: 'in',
                    value: ['标准型'],
                  },
                ],
              },
              page: { limit: 0, count: 0 },
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const updateRemark = (remark: string) => {
      recycleForm.value.remark = remark;
    };
    const updateHosts = (hosts: any[]) => {
      tableHosts.value = hosts;
    };
    const updateSelectedHosts = (hosts: any[]) => {
      tableSelectedHosts.value = hosts;
      recycleForm.value.ips = hosts.map((item) => item.ip);
    };
    const updateTypes = (returnPlan: { value: any }) => {
      returnPlan.value = returnPlan;
    };
    const handleNext = () => {
      const { cvm, pm, skipConfirm } = returnPlan.value;
      if (active.value === 1 && (!cvm || !pm || skipConfirm === '')) {
        // $refs.resourceType?.$refs?.recycleForm?.validate();
        return;
      }
      active.value += 1;
    };
    const upDrawer = (val: boolean) => {
      drawer.value = val;
    };

    watch(
      () => bkBizId.value,
      (val) => {
        if (val) {
          getListData();
        }
      },
    );
    const renderButton = () => {
      if (active.value === 1) return null;

      return (
        <bk-button
          size='small'
          class='mr10'
          onClick={() => {
            active.value -= 1;
          }}>
          上一步
        </bk-button>
      );
    };
    return () => (
      <div class='div-RecyclingResources'>
        <div class='div-title'>回收资源</div>
        <div class='div-components'>
          <bk-steps class='div-steps' cur-step={active.value} steps={objectSteps.value} />
          {active.value === 1 && (
            <ResourceSelect
              class='div-ResourceSelect'
              table-hosts={tableHosts.value}
              table-selected-hosts={tableSelectedHosts.value}
              remark={recycleForm.value.remark}
              onUpdateHosts={updateHosts}
              onDrawer={upDrawer}
              onUpdateSelectedHosts={updateSelectedHosts}
              onUpdateRemark={updateRemark}></ResourceSelect>
          )}
          {active.value === 2 && (
            <ResourceType ref='resourceType' return-plan={returnPlan} updateTypes={updateTypes}></ResourceType>
          )}
          {active.value === 3 && (
            <ResourceConfirm recycle-form={recycleForm} return-plan={returnPlan}></ResourceConfirm>
          )}
        </div>
        <div class='div-Button'>
          {renderButton()}
          {active.value < 3 && (
            <bk-button
              theme='primary'
              class='mr10'
              size='small'
              disabled={active.value === 0 && !tableSelectedHosts.value.length}
              onClick={handleNext}>
              下一步
            </bk-button>
          )}
          {active.value === 3 && (
            <span class='ml-10'>
              <bk-button
                theme='primary'
                size='small'
                disabled={!selectedHosts.value.length}
                onClick={() => {
                  dialogVisible.value = true;
                }}>
                提 交
              </bk-button>
            </span>
          )}
        </div>
        <Dialog title='注意' is-show={dialogVisible.value} custom-class='notice' width='520px'>
          <p>
            1. 销毁后所有数据<span class='main'>将被清除且不可恢复</span>，CVM会同时<span class='main'>销毁</span>
            挂载在实例上的包年包月数据盘；
          </p>
          <p>
            2. 非立即销毁, 隔离期间费用会<span class='main'>继续核算至业务下</span>；
          </p>
          <p>
            3. 计划外回收，公司会核算给回收<span class='main'>业务35天的滞留成本</span>；
          </p>
          <p>
            4. CVM存量机型回收后，公司会核算给回收<span class='main'>业务20%的成本</span>；
          </p>
          <p>
            5. 物理机未退役设备的回收，回收业务需要<span class='main'>承担60%的滞留成本</span>；
          </p>
          <p>
            6. <span class='main'>请提前录入 </span>
            <a href='https://yunti.woa.com/plans/return' target='_blank'>
              回收计划
            </a>
          </p>
          <p>
            7. 更多信息，请查看
            <a href='https://yunti.woa.com/news/15' target='_blank'>
              公司资源退回管理策略
            </a>
          </p>
          <br />
          {/* <el-checkbox v-model="checked">
        我已知悉以上须知内容和风险
      </el-checkbox> */}
          {/* <span slot='footer' class='dialog-footer'> */}
          {/* <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button
          type="primary"
          :disabled="!checked"
          @click="handleConfirm"
        >确 定</el-button> */}
          {/* </span> */}
        </Dialog>
        <CommonSideslider v-model:isShow={drawer.value} title='选择服务器' width={1150}>
          <Tab v-model:active={activetab.value} type='unborder-card'>
            <BkTabPanel key={0} name={0} label='根据业务选择(单业务回收场景)'>
              <BSCommonTable>
                {{
                  tabselect: () => (
                    <>
                      <div class='displayflex'>
                        <div class='mr-10'>业务</div>
                        <bk-select class='tbkselect'>
                          {[].map((item) => (
                            <bk-option
                              key={item.require_type}
                              value={item.require_name}
                              label={item.require_type}></bk-option>
                          ))}
                        </bk-select>
                      </div>

                      {bkBizId.value && <span> / 空闲机池 / 待回收</span>}
                    </>
                  ),
                }}
              </BSCommonTable>
            </BkTabPanel>
            <BkTabPanel key={1} name={1} label='手动输入(多业务回收场景)'>
              <bk-input
                type='textarea'
                v-model={lips.value}
                text
                placeholder='请输入 IP地址/固资号，多个换行分割，最多支持500个'
                rows={1}></bk-input>
            </BkTabPanel>
          </Tab>
        </CommonSideslider>
        <CommonSideslider v-model:isShow={ResourcesTotal.value} title='互娱作业管理平台ijabs' width={1150}>
          <RTCommonTable></RTCommonTable>
        </CommonSideslider>
      </div>
    );
  },
});
