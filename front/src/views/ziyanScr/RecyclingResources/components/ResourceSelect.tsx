import { defineComponent, ref, computed, watch } from 'vue';
import './index.scss';
import apiService from '../../../../api/scrApi';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  name: 'ResourceSelect',
  props: {
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
    const count = ref(0);
    const TableData = ref([]);
    const allRecycleHostIps = computed(() => {
      return props.tableHosts.map((item) => item.ip).join('\n');
    });
    const recycleFailedHostIps = computed(() => {
      return props.tableHosts
        .map((item) => {
          if (!item.recyclable) {
            return item.ip;
          }
          return null;
        })
        .join('\n');
    });
    const localRemark = ref('');

    watch(
      () => localRemark.value,
      (val) => {
        emit('updateRemark', val);
      },
    );
    /** 刷新可回收状态 */
    const refresh = async () => {
      const ips = props.tableHosts.map((item) => item.ip);
      const { info } = await apiService.getRecyclableHosts({
        ips,
      });

      const list = info || [];
      emit('updateHosts', list);
      //   this.$message.success('刷新成功');
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
        <div class='CommonTable'>
          <bk-button class='bk-button' theme='primary' onClick={handleCleardrawer}>
            选择服务器
          </bk-button>
          <bk-button class='bk-button' onClick={refresh}>
            刷新状态
          </bk-button>
          <bk-button
            class='bk-button'
            v-clipboard={allRecycleHostIps.value}
            disabled={allRecycleHostIps.value.length === 0}>
            复制所有IP{count.value}
          </bk-button>
          <bk-button
            class='bk-button'
            theme='danger'
            v-clipboard={recycleFailedHostIps.value}
            disabled={recycleFailedHostIps.value.length === 0}>
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
        <bk-table align='left' row-hover='auto' columns={RRcolumns} data={TableData.value} show-overflow-tooltip />
      </div>
    );
  },
});
