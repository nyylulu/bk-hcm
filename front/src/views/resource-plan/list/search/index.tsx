import { defineComponent, onBeforeMount, ref } from 'vue';
import Panel from '@/components/panel';
import { Input, Select, Button, DatePicker, Message } from 'bkui-vue';
import cssModule from './index.module.scss';
import MemberSelect from '@/components/MemberSelect';
import { useResourcePlanStore } from '@/store';
import BusinessSelector from '@/components/business-selector/index.vue';
import { timeFormatter } from '@/common/util';
import type { IListTicketsParam } from '@/typings/resourcePlan';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  emits: ['search'],

  setup(_, { emit }) {
    const { Option } = Select;
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const initialSearchModel: Partial<IListTicketsParam> = {
      bk_biz_ids: [],
      statuses: [],
      ticket_ids: [],
      applicants: [],
      submit_time_range: undefined,
    };
    const statues = ref<{ status: string; status_name: string }[]>([]);
    const searchModel = ref(JSON.parse(JSON.stringify(initialSearchModel)));

    const getStatues = async () => {
      try {
        const res = await resourcePlanStore.getStatusList();
        statues.value = res?.data?.details || [];
      } catch (error: any) {
        Message({
          message: error?.message || error,
          theme: 'error',
        });
      }
    };

    const handleSearch = () => {
      emit('search', searchModel.value);
    };

    const handleReset = () => {
      searchModel.value = JSON.parse(JSON.stringify(initialSearchModel));
      emit('search', undefined);
    };

    const handleInputTicket = (val: string) => {
      searchModel.value.ticket_ids = val.split(';').filter((v) => v);
    };

    const handleChangeDate = (key: string, val: string[]) => {
      if (val[0] && val[1]) {
        searchModel.value[key] = {
          start: timeFormatter(val[0], 'YYYY-MM-DD'),
          end: timeFormatter(val[1], 'YYYY-MM-DD'),
        };
      } else {
        searchModel.value[key] = undefined;
      }
    };

    onBeforeMount(getStatues);

    return () => (
      <Panel class={cssModule['mb-16']}>
        <div class={cssModule['search-grid']}>
          <div>
            <div class={cssModule['search-label']}>{t('业务')}</div>
            <BusinessSelector
              v-model={searchModel.value.bk_biz_ids}
              multiple={true}
              authed={true}
              autoSelect={true}
              isShowAll={true}
            />
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('单据状态')}</div>
            <Select multiple v-model={searchModel.value.statuses}>
              {statues.value.map((item) => (
                <Option key={item.status} name={item.status_name} id={item.status}>
                  {item.status_name}
                </Option>
              ))}
            </Select>
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('预测单号')}</div>
            <Input
              modelValue={searchModel.value.ticket_ids.join(';')}
              placeholder={t('请输入预测单号，多个预测单号可使用分号分隔')}
              onChange={handleInputTicket}
            />
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('提单人')}</div>
            <MemberSelect v-model={searchModel.value.applicants} />
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('提单时间')}</div>
            <DatePicker
              modelValue={[searchModel.value.submit_time_range?.start, searchModel.value.submit_time_range?.end]}
              onChange={(val: string[]) => handleChangeDate('submit_time_range', val)}
              type='daterange'
            />
          </div>
        </div>
        <Button theme='primary' class={cssModule['search-button']} onClick={handleSearch}>
          {t('查询')}
        </Button>
        <Button onClick={handleReset} class={cssModule['search-button']}>
          {t('重置')}
        </Button>
      </Panel>
    );
  },
});
