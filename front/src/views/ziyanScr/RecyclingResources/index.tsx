import { defineComponent, ref, watch, onMounted } from 'vue';
import ResourceSelect from './components/ResourceSelect';
import ResourceType from './components/ResourceType';
import ResourceConfirm from './components/ResourceConfirm';
import { Dialog, Tab, Button } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '@/components/common-sideslider';
import { useAccountStore } from '@/store';
import apiService from '@/api/scrApi';
import { useRouter } from 'vue-router';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
export default defineComponent({
  name: 'RecyclingResources',
  setup() {
    const accountStore = useAccountStore();
    const active = ref(1);
    const router = useRouter();
    const objectSteps = ref([{ title: '输入IP/固资' }, { title: '确认回收类型' }, { title: '信息确认与提交' }]);
    const tableHosts = ref([]);
    const bkBizId = ref();
    const checked = ref(false);
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
      skipConfirm: true,
    });
    const drawerSelectedHosts = ref([]);
    const drawer = ref(false);
    const ResourcesTotal = ref(false);
    const lips = ref();
    const activetab = ref(0);
    const serverTableData = ref([]);
    const businessList = ref([]);
    const { columns: BScolumns } = useColumns('BusinessSelection');
    const { columns: RTcolumns } = useColumns('ResourcesTotal');
    const { CommonTable: RTCommonTable } = useTable({
      tableOptions: {
        columns: RTcolumns,
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          payload: {
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
    const updateConfirm = (selectedList: any[]) => {
      selectedHosts.value = selectedList;
    };
    const getBusinessesList = async () => {
      const { data } = await accountStore.getBizListWithAuth();
      businessList.value = data;
    };
    const handleNext = () => {
      const { cvm, pm } = returnPlan.value;
      if (active.value === 1 && (!cvm || !pm)) {
        // $refs.resourceType?.$refs?.recycleForm?.validate();
        return;
      }
      active.value += 1;
    };
    const upDrawer = (val: boolean) => {
      drawer.value = val;
    };
    const page = ref({
      start: 0,
      limit: 10,
      enable_count: false,
    });
    const initPage = () => {
      page.value.start = 0;
      page.value.limit = 10;
      page.value.enable_count = false;
    };
    const total = ref(0);
    const getListData = async (getCount?: boolean) => {
      let pageObj = page.value;
      if (getCount) {
        pageObj = {
          start: 0,
          limit: 0,
          enable_count: true,
        };
      }
      const data = await apiService.getRecycleList({
        bkBizId: bkBizId.value,
        page: pageObj,
      });
      if (getCount) {
        total.value = data?.count;
      } else {
        serverTableData.value = data?.info || [];
      }
    };
    watch(
      () => bkBizId.value,
      (val) => {
        if (val) {
          initPage();
          getListData(true);
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
    /** 更新选中的资源 */
    const updateTableHosts = (hosts: any) => {
      const obj = {};
      tableHosts.value = tableHosts.value.concat(hosts).reduce((prev, cur) => {
        if (!obj[cur.ip]) {
          obj[cur.ip] = true;
          cur.recyclable ? prev.push(cur) : prev.unshift(cur);
        }
        return prev;
      }, []);
    };
    const handleSubmit = async () => {
      if (activetab.value === 1) {
        const { info } = await apiService.getRecyclableHosts({
          ips: drawerSelectedHosts.value,
        });

        const hosts = info || [];
        updateTableHosts(hosts);
        drawer.value = false;
      } else {
        checkHostRecyclableStatus();
      }
    };
    /** 手动输入查询可回收状态 */
    const checkHostRecyclableStatus = async () => {
      const ipv4 = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
      const ipv6 = /^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}$/;
      const ips: any[] = [];
      const assetIds: any[] = [];
      const ipsList: any[] = [];
      if (ipsList.length > 500) {
        // this.$message.error(`最多添加500台主机,请删除${ipsList.length - 500}台后重试`)
        return;
      }
      ipsList.forEach((item) => {
        if (ipv4.test(item) || ipv6.test(item)) {
          ips.push(item);
        } else {
          assetIds.push(item);
        }
      });
      const { info } = await apiService.getRecyclableHosts({
        ips,
        asset_ids: assetIds,
      });
      const drawerHosts = info || [];
      updateTableHosts(drawerHosts);
      drawer.value = false;
    };
    const handleConfirm = async () => {
      const orderId = selectedHosts.value.map((item) => item.order_id);
      const { result } = await apiService.startRecycleList(orderId);
      if (result) {
        takeSnapshot();
        router.push({
          path: '/ziyanScr/hostRecycling/resources',
        });
      }
      checked.value = false;
      dialogVisible.value = false;
    };
    const takeSnapshot = () => {
      recycleForm.value = {
        ips: [],
        remark: '',
      };
    };
    const triggerShow = (val: boolean) => {
      checked.value = val;
      dialogVisible.value = val;
    };
    onMounted(() => {
      getBusinessesList();
    });
    return () => (
      <div class='div-RecyclingResources'>
        <DetailHeader>
          <span class='header-title-prefix'>主机回收</span>
        </DetailHeader>
        <div class='common-sub-main-container'>
          <div class='sub-main-content'>
            <div class='div-title'>回收资源</div>
            <div class='div-components'>
              <bk-steps class='div-steps' cur-step={active.value} steps={objectSteps.value} />
              {active.value === 1 && (
                <ResourceSelect
                  class='div-ResourceSelect'
                  table-hosts={tableHosts.value}
                  table-selected-hosts={tableSelectedHosts.value}
                  onUpdateHosts={updateHosts}
                  onDrawer={upDrawer}
                  onUpdateSelectedHosts={updateSelectedHosts}
                  onUpdateRemark={updateRemark}></ResourceSelect>
              )}
              {active.value === 2 && (
                <ResourceType ref='resourceType' returnPlan={returnPlan.value} updateTypes={updateTypes}></ResourceType>
              )}
              {active.value === 3 && (
                <ResourceConfirm
                  recycleForm={recycleForm.value}
                  returnPlan={returnPlan.value}
                  onUpdateConfirm={updateConfirm}></ResourceConfirm>
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
                    // disabled={!selectedHosts.value.length}
                    onClick={() => {
                      dialogVisible.value = true;
                    }}>
                    提 交
                  </bk-button>
                </span>
              )}
            </div>
          </div>
        </div>
        <Dialog title='注意' is-show={dialogVisible.value} custom-class='notice' width='520px'>
          {{
            default: () => (
              <>
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
                <bk-checkbox v-model={checked.value}>我已知悉以上须知内容和风险</bk-checkbox>
              </>
            ),
            footer: () => (
              <>
                <Button theme='primary' onClick={handleConfirm} disabled={!checked.value}>
                  确定
                </Button>
                <Button class='dialog-cancel' onClick={() => triggerShow(false)}>
                  取消
                </Button>
              </>
            ),
          }}
        </Dialog>
        <CommonSideslider v-model:isShow={drawer.value} title='选择服务器' width={1150} onHandleSubmit={handleSubmit}>
          <Tab v-model:active={activetab.value} type='unborder-card'>
            <BkTabPanel key={0} name={0} label='根据业务选择(单业务回收场景)'>
              <div class='displayflex' style='margin: 20px 0;'>
                <div class='displayflex'>
                  <div class='mr-10'>业务</div>
                  <bk-select class='item-warp-component' v-model={bkBizId.value}>
                    {businessList.value.map((item) => (
                      <bk-option key={item.id} value={item.id} label={item.name}></bk-option>
                    ))}
                  </bk-select>
                </div>
                {bkBizId.value && <span style='width:520px'> / 空闲机池 / 待回收</span>}
              </div>

              <bk-table
                align='left'
                row-hover='auto'
                columns={BScolumns}
                data={serverTableData.value}
                show-overflow-tooltip
              />
              <bk-pagination
                style='float: right;margin: 20px 0;'
                v-model={page.value.start}
                count={total.value}
                limit={page.value.limit}
              />
            </BkTabPanel>
            <BkTabPanel key={1} name={1} label='手动输入(多业务回收场景)'>
              <bk-input
                type='textarea'
                style='width:520px'
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
