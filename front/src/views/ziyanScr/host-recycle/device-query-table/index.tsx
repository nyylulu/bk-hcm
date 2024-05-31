import { defineComponent, ref, onMounted } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
// import {
//   getRecycleHosts,
//   getDeviceTypeList,
//   getRegionList,
//   getZoneList,
//   getRecycleStageOpts,
// } from '@/api/host/recycle';
import { Search } from 'bkui-vue/lib/icon';
export default defineComponent({
  setup() {
    const defaultDeviceForm = {
      bkBizId: [],
      orderId: [],
      suborderId: [],
      deviceType: [],
      bkZoneName: [],
      subZone: [],
      stage: [],
      ip: [],
      handleTime: [new Date(), new Date()],
      // start: getDate('yyyy-MM-dd', -30),
      // end: getDate('yyyy-MM-dd', 0),
      bkUsername: '',
      // bkUsername: [this.$store.getters.name],
    };
    const deviceForm = ref(defaultDeviceForm);
    const bussinessList = [];
    const deviceTypeList = ref([]);
    const bkZoneNameList = ref([]);
    const subZoneList = ref([]);
    const stageList = ref([]);
    const recycleMen = [];
    const { columns } = useColumns('deviceQuery');
    const tableColumns = [...columns];
    const pageInfo = ref({
      start: 0,
      limit: 10,
      total: 0,
    });
    const { CommonTable } = useTable({
      tableOptions: {
        columns: tableColumns,
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
              page: [],
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const deviceList = ref([]);
    const getDevicelist = (enableCount = false) => {
      return enableCount;
      // deviceForm.value.orderId = deviceForm.value.orderId.map((item) => Number(item));
      // getRecycleHosts(
      //   {
      //     ...deviceForm.value,
      //     ...pageInfo.value,
      //   },
      //   {
      //     //   requestId,
      //     enableCount,
      //   },
      // ).then((res) => {
      //   if (enableCount) pageInfo.value.total = res.data?.count;
      //   deviceList.value = res.data?.info || [];
      // });
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      getDevicelist(true);
    };
    const clearFilter = () => {
      deviceForm.value = defaultDeviceForm;
      filterOrders();
    };
    // const fetchDeviceTypeList = () => {
    //   getDeviceTypeList().then((res) => {
    //     deviceTypeList.value = res.data.info || [];
    //   });
    // };
    // const fetchRegionList = () => {
    //   getRegionList().then((res) => {
    //     bkZoneNameList.value = res.data.info || [];
    //   });
    // };
    // const fetchZoneList = () => {
    //   getZoneList().then((res) => {
    //     subZoneList.value = res.data.info || [];
    //   });
    // };
    // const fetchStageList = () => {
    //   getRecycleStageOpts().then((res) => {
    //     stageList.value = res.data.info || [];
    //   });
    // };
    onMounted(() => {
      //   getDevicelist(true);
      //   fetchDeviceTypeList();
      //   fetchRegionList();
      //   fetchZoneList();
      //   fetchStageList();
    });
    return () => (
      <div>
        <CommonTable>
          {{
            tabselect: () => (
              <bk-form label-width='110' class='bill-filter-form' model={deviceForm}>
                <bk-form-item label='业务'>
                  {/* <AppSelect>
                    {
                      {
                        //   append: () => (
                        //     <div class={'app-action-content'}>
                        //       <i class={'hcm-icon bkhcm-icon-plus-circle app-action-content-icon'} />
                        //       <span class={'app-action-content-text'}>新建业务</span>
                        //     </div>
                        //   ),
                      }
                    }
                  </AppSelect> */}
                  {/* TODO 新AppSelect 旧cr-biz-select */}
                  <bk-select v-model={deviceForm.value.bkBizId} multiple clearable placeholder='请选择业务'>
                    {bussinessList.map(({ key, value }) => {
                      return <bk-option key={key} label={value} value={key}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='单号'>
                  {/* TODO 旧float-input*/}
                  <bk-input v-model={deviceForm.value.orderId} />
                </bk-form-item>
                <bk-form-item label='子单号'>
                  {/* TODO 旧float-input*/}
                  <bk-input v-model={deviceForm.value.suborderId} />
                </bk-form-item>
                <bk-form-item label='机型'>
                  <bk-select v-model={deviceForm.value.deviceType} multiple clearable placeholder='请选择机型'>
                    {deviceTypeList.value.map((item) => {
                      return <bk-option key={item} label={item} value={item}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='地域'>
                  <bk-select v-model={deviceForm.value.bkZoneName} multiple clearable placeholder='请选择地域'>
                    {bkZoneNameList.value.map((item) => {
                      return <bk-option key={item} label={item} value={item}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='园区'>
                  <bk-select v-model={deviceForm.value.subZone} multiple clearable placeholder='请选择园区'>
                    {subZoneList.value.map((item) => {
                      return <bk-option key={item} label={item} value={item}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='状态'>
                  <bk-select v-model={deviceForm.value.stage} multiple clearable placeholder='请选择状态'>
                    {stageList.value.map(({ stage, description }) => {
                      return <bk-option key={stage} label={description} value={stage}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='回收IP'>
                  {/* TODO 旧float-input*/}
                  <bk-input v-model={deviceForm.value.ip} />
                </bk-form-item>
                <bk-form-item label='回收人'>
                  {/* TODO MemberSelect使用不对 */}
                  {/* <member-select v-model={deviceForm.value.bkUsername} multiple clearable allowCreate /> */}
                  <bk-select v-model={deviceForm.value.bkUsername} multiple clearable placeholder='请选择业务'>
                    {recycleMen.map(({ key, value }) => {
                      return <bk-option key={key} label={value} value={key}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='完成时间'>
                  {/* TODO 是否封装成旧的 */}
                  <bk-date-picker v-model={deviceForm.value.handleTime} type='daterange' />
                </bk-form-item>
                <bk-form-item class='bill-form-btn' label-width='20'>
                  <bk-button
                    theme='primary'
                    //   :loading="$isLoading(orders.requestId)"
                    native-type='submit'
                    onClick={filterOrders}>
                    <Search />
                    查询
                  </bk-button>
                  <bk-button
                    //   :loading="$isLoading(orders.requestId)"
                    onClick={clearFilter}>
                    {/* TODO icon='el-icon-refresh' */}
                    <Search />
                    清空
                  </bk-button>
                  <export-to-excel-button data={deviceList.value} columns={tableColumns} filename='回收设备列表' />
                </bk-form-item>
              </bk-form>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
