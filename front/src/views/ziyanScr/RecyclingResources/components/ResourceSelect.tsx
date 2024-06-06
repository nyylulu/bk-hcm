import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import apiService from '../../../../api/scrApi';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  name: 'ResourceSelect',
  props: {
    remark: {
      type: String,
      default: '',
    },
    tableHosts: {
      type: Array,
      default: () => [],
    },
    tableSelectedHosts: {
      type: Array,
      default: () => [],
    },
    updateRemark: {
      type: Function,
      default: () => {},
    },
    updateHosts: {
      type: Function,
      default: () => {},
    },
    updateSelectedHosts: {
      type: Function,
      default: () => {},
    },
  },
  emits: ['updateHosts', 'updateSelectedHosts', 'updateRemark', 'Drawer'],
  setup(props, { emit }) {
    const { columns: RRcolumns } = useColumns('RecyclingResources');

    const { CommonTable: RRCommonTable } = useTable({
      tableOptions: {
        columns: RRcolumns,
      },
      requestOption: {
        dataPath: 'data.info',
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

    const count = ref(0);
    const allRecycleHostIps = ref([]);
    const recycleFailedHostIps = ref([]);
    const localRemark = ref('');

    watch(
      () => localRemark.value,
      (val) => {
        emit('updateRemark', val);
      },
    );
    /** 刷新可回收状态 */
    const refresh = () => {
      const ips = props.tableHosts.map((item) => item.ip);
      apiService
        .getRecyclableHosts({
          ips,
        })
        .then((res) => {
          const list = res.data?.info || [];
          emit('updateHosts', list);
          //   this.$message.success('刷新成功');
        });
    };
    /** 清空列表 */
    const handleClear = () => {
      emit('updateHosts', []);
      emit('updateSelectedHosts', []);
    };
    const handleCleardrawer = () => {
      emit('Drawer', true);
    };
    return () => (
      <div>
        <bk-alert theme='warning' show-icon={false}>
          {{
            title: () => (
              <>
                <div class='alerttext fw'>上交机器的检测条件</div>
                <div class='alerttext'>1、需要有业务的回收资源权限</div>
                <div class='alerttext'>2、设备必须在业务的”空闲机“-”待回收“模块</div>
                <div class='alerttext'>3、回收人是设备的主责任人或者备份负责人</div>
              </>
            ),
          }}
        </bk-alert>
        <RRCommonTable>
          {{
            tabselect: () => (
              <div class='CommonTable'>
                <bk-button class='bk-button' theme='primary' onClick={handleCleardrawer}>
                  选择服务器
                </bk-button>
                <bk-button class='bk-button' onClick={refresh}>
                  刷新状态
                </bk-button>
                <bk-button class='bk-button' disabled={allRecycleHostIps.value.length === 0}>
                  复制所有IP{count.value}
                </bk-button>
                <bk-button class='bk-button' theme='danger' disabled={recycleFailedHostIps.value.length === 0}>
                  复制不可回收IP{count.value}
                </bk-button>
                <bk-button class='bk-button' theme='primary' onClick={handleClear}>
                  清空列表
                </bk-button>
                <div class='displayflex'>
                  <div class='displayflex-test'>备注</div>
                  <bk-input
                    type='textarea'
                    v-model={localRemark.value}
                    text
                    placeholder='回收备注,256 字以内'
                    rows={1}
                    maxlength={255}></bk-input>
                </div>
              </div>
            ),
          }}
        </RRCommonTable>
      </div>
    );
  },
});
