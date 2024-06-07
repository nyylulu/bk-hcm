import { defineComponent, onBeforeMount, reactive, ref } from 'vue';
import Panel from '@/components/panel';
import { Input, Select, Button, DatePicker } from 'bkui-vue';
import cssModule from '../index.module.scss';
import MemberSelect from '@/components/MemberSelect';
import { useResourcePlanStore } from '@/store';
import dayjs from 'dayjs';
import BusinessSelector from '@/components/business-selector/index.vue';

export default defineComponent({
  setup(props, { expose }) {
    const { Option } = Select;
    const resourcePlanStore = useResourcePlanStore();

    const initialSearchFormModel = {
      bk_biz_ids: undefined,
      obs_projects: undefined,
      ticket_ids: undefined,
      applicants: undefined,
      submit_time_range: undefined,
    };

    const projectList = ref([]);
    const searchData = ref({});
    const searchFormModel = reactive({ ...initialSearchFormModel });

    const getAllSearchList = async () => {
      try {
        const res = await resourcePlanStore.getObsProjects();
        projectList.value = res.data;
      } catch (error) {
        console.error(error, 'error'); // eslint-disable-line no-console
      }
    };

    const handleSearch = () => {
      const { submit_time_range } = searchFormModel;
      const isTime = !!searchFormModel.submit_time_range?.length;
      searchData.value = {
        ...searchFormModel,
        submit_time_range: isTime
          ? {
              start: formatTime(submit_time_range[0]),
              end: formatTime(submit_time_range[1]),
            }
          : undefined,
      };
    };

    const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD');

    const handleReset = () => {
      Object.assign(searchFormModel, { ...initialSearchFormModel });
      searchData.value = {};
    };

    const handleClear = (type: string) => {
      searchFormModel[type] = undefined;
      searchData.value[type] = undefined;
    };

    onBeforeMount(getAllSearchList);

    expose({
      searchData,
    });
    return () => (
      <Panel class={cssModule['mb-16']}>
        <div class={cssModule['search-grid']}>
          <div>
            <div class={cssModule['search-label']}>业务</div>
            <BusinessSelector
              v-model={searchFormModel.bk_biz_ids}
              authed={true}
              auto-select={true}
              onHandleClear={() => handleClear('bk_biz_ids')}></BusinessSelector>
          </div>
          <div>
            <div class={cssModule['search-label']}>项目类型</div>
            <Select
              onClear={() => handleClear('obs_projects')}
              placeholder='请选择'
              multiple
              v-model={searchFormModel.obs_projects}>
              {projectList.value.map((item) => (
                <Option key={item} label={item} value={item}>
                  {item}
                </Option>
              ))}
            </Select>
          </div>
          <div>
            <div class={cssModule['search-label']}>预测单号</div>
            <Input placeholder='请选择' v-model={searchFormModel.ticket_ids}></Input>
          </div>
          <div>
            <div class={cssModule['search-label']}>提单人</div>
            <MemberSelect class={cssModule['w-full']} v-model={searchFormModel.applicants} />
          </div>
          <div>
            <div class={cssModule['search-label']}>提单时间</div>
            <DatePicker
              class={cssModule['w-full']}
              v-model={searchFormModel.submit_time_range}
              placeholder='请选择'
              type='daterange'></DatePicker>
          </div>
        </div>
        <Button theme='primary' class={cssModule['mt-24']} onClick={() => handleSearch()}>
          查询
        </Button>
        <Button onClick={() => handleReset()} class={[cssModule['mt-24'], cssModule['ml-8']]}>
          重置
        </Button>
      </Panel>
    );
  },
});
