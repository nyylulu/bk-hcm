import { defineComponent, onMounted, reactive, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { Button, DatePicker, Select, Tab, TagInput } from 'bkui-vue';
import { Plus, Search } from 'bkui-vue/lib/icon';
import MemberSelect from '@/components/MemberSelect';
import { useZiyanScrStore } from '@/store';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { cleanPayload, getDate } from '@/utils';
import { timeFormatter } from '@/common/util';
import './index.scss';

const { TabPanel } = Tab;

interface ResourceManageFilterType {
  id?: string[];
  bk_username?: string[];
  phase?: string[];
  start: string;
  end: string;
}

/**
 * SCR - 资源上下架
 */
export default defineComponent({
  name: 'ScrResourceManage',
  setup() {
    const router = useRouter();
    const route = useRoute();
    const ziyanScrStore = useZiyanScrStore();

    const { columns: scrResourceOnlineColumns } = useColumns('scrResourceOnline');
    const { columns: scrResourceOfflineColumns } = useColumns('scrResourceOffline');

    const {
      CommonTable: ScrResourceOnlineTable,
      getListData: reloadScrResourceOnlineTable,
      pagination: scrResourceOnlinePagination,
    } = useTable({
      tableOptions: {
        columns: scrResourceOnlineColumns,
      },
      scrConfig: () => ({
        url: '/api/v1/woa/pool/findmany/launch/task',
        payload: {
          ...cleanPayload(filter),
          id: filter.id.length ? filter.id.map((id) => Number(id)) : undefined,
          start: timeFormatter(filter.start, 'YYYY-MM-DD'),
          end: timeFormatter(filter.end, 'YYYY-MM-DD'),
        },
      }),
    });

    const {
      CommonTable: ScrResourceOfflineTable,
      getListData: reloadScrResourceOfflineTable,
      pagination: scrResourceOfflinePagination,
    } = useTable({
      tableOptions: {
        columns: scrResourceOfflineColumns,
      },
      scrConfig: () => ({
        url: '/api/v1/woa/pool/findmany/recall/task',
        payload: {
          ...cleanPayload(filter),
          id: filter.id.length ? filter.id.map((id) => Number(id)) : undefined,
          start: timeFormatter(filter.start, 'YYYY-MM-DD'),
          end: timeFormatter(filter.end, 'YYYY-MM-DD'),
        },
      }),
    });

    const reloadDataList = () => {
      if (activeType.value === 'online') {
        scrResourceOnlinePagination.start = 0;
        reloadScrResourceOnlineTable();
      } else {
        scrResourceOfflinePagination.start = 0;
        reloadScrResourceOfflineTable();
      }
    };

    const types = [
      {
        label: '资源上架',
        value: 'online',
        Component: ScrResourceOnlineTable,
      },
      {
        label: '资源下架',
        value: 'offline',
        Component: ScrResourceOfflineTable,
      },
    ];

    const activeType = ref(route.query.type || 'online');

    watch(activeType, (val) => {
      router.push({ query: { type: val } });
      reloadDataList();
    });

    const getDefaultFilter = (): ResourceManageFilterType => ({
      id: [],
      bk_username: [],
      phase: [],
      start: getDate('yyyy-MM-dd', -30),
      end: getDate('yyyy-MM-dd', 0),
    });
    const filter = reactive(getDefaultFilter());

    const phaseList = ref([]);
    onMounted(() => {
      // 组件挂载完成后，请求单据状态list
      const getPhaseList = async () => {
        const res = await ziyanScrStore.getTaskStatusList();
        phaseList.value = res.data.info || [];
      };

      getPhaseList();
      reloadDataList();
    });

    const clearFilter = () => {
      Object.assign(filter, getDefaultFilter());
      reloadDataList();
    };

    return () => (
      <div class='scr-resource-manage-page'>
        <Tab v-model:active={activeType.value} type='card-grid'>
          {types.map(({ label, value, Component }) => (
            <TabPanel key={value} label={label} name={value} renderDirective='if'>
              <div class='manage-container'>
                <div class='filter-container'>
                  <div class='filter-item mr8'>
                    <Button theme='primary'>
                      <Plus /> 发起{activeType.value === 'online' ? '上架' : '下架'}
                    </Button>
                  </div>
                  <div class='filter-item mr8'>
                    <span class='mr8'>单号</span>
                    <TagInput
                      v-model={filter.id}
                      class='w200'
                      allow-create
                      collapse-tags
                      createTagValidator={(tag) => /^[1-9]\d*$/.test(tag)}
                    />
                  </div>
                  <div class='filter-item mr8'>
                    <span class='mr8'>创建人</span>
                    <MemberSelect class='w200' v-model={filter.bk_username} />
                  </div>
                  <div class='filter-item mr8'>
                    <span class='mr8'>创建时间</span>
                    <DatePicker class='w150' v-model={filter.start} />
                    <span class='m4'>-</span>
                    <DatePicker class='w150' v-model={filter.end} />
                  </div>
                  <div class='filter-item mr8'>
                    <span class='mr8'>单据状态</span>
                    <Select v-model={filter.phase} multiple>
                      {phaseList.value.map(({ description, status }) => (
                        <Select.Option key={status} id={status} name={description} />
                      ))}
                    </Select>
                  </div>
                  <div class='filter-item mr8'>
                    <Button onClick={reloadDataList}>
                      <Search />
                      查询
                    </Button>
                  </div>
                  <div class='filter-item'>
                    <Button onClick={clearFilter}>清空</Button>
                  </div>
                </div>
                <Component />
              </div>
            </TabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
