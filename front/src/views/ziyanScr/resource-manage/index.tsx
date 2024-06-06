import { defineComponent, onMounted, reactive, ref, watchEffect } from 'vue';
import { useRouter } from 'vue-router';
import { Button, DatePicker, Select, Tab, TagInput } from 'bkui-vue';
import { Plus, Search } from 'bkui-vue/lib/icon';
import MemberSelect from '@/components/MemberSelect';
import RemoteTable from '@/components/RemoteTable';
import { useZiyanScrStore } from '@/store';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
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
    const ziyanScrStore = useZiyanScrStore();

    const { columns: scrResourceOnlineColumns } = useColumns('scrResourceOnline');
    const { columns: scrResourceOfflineColumns } = useColumns('scrResourceOffline');

    const types = [
      {
        label: '资源上架',
        value: 'online',
        columns: scrResourceOnlineColumns,
        url: '/api/v1/woa/pool/findmany/launch/task',
      },
      {
        label: '资源下架',
        value: 'offline',
        columns: scrResourceOfflineColumns,
        url: '/api/v1/woa/pool/findmany/recall/task',
      },
    ];

    const activeType = ref('online');
    watchEffect(() => {
      router.push({ query: { type: activeType.value } });
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
    });

    const remoteTableRef = ref();
    const filterList = () => {
      remoteTableRef.value.pagination.start = 0;
      remoteTableRef.value.getDataList();
    };
    const clearFilter = () => {
      Object.assign(filter, getDefaultFilter());
      filterList();
    };

    return () => (
      <div class='scr-resource-manage-page'>
        <Tab v-model:active={activeType.value} type='card-grid'>
          {types.map(({ label, value, columns, url }) => (
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
                    <Button onClick={filterList}>
                      <Search />
                      查询
                    </Button>
                  </div>
                  <div class='filter-item'>
                    <Button onClick={clearFilter}>清空</Button>
                  </div>
                </div>
                <RemoteTable
                  ref={remoteTableRef}
                  columns={columns}
                  noSort
                  path={{ start: 'start', limit: 'limit', count: 'enable_count', data: 'info', total: 'count' }}
                  apis={[
                    {
                      url,
                      payload: () => ({
                        ...cleanPayload(filter),
                        id: filter.id.length ? filter.id.map((id) => Number(id)) : undefined,
                        start: timeFormatter(filter.start, 'YYYY-MM-DD'),
                        end: timeFormatter(filter.end, 'YYYY-MM-DD'),
                        filter: undefined,
                      }),
                    },
                  ]}
                />
              </div>
            </TabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
