import { defineComponent, onBeforeMount, ref } from 'vue';
import Panel from '@/components/panel';
import { Input, Select, Button, DatePicker, Message } from 'bkui-vue';
import cssModule from './index.module.scss';
import MemberSelect from '@/components/MemberSelect';
import { useResourcePlanStore } from '@/store';
import BusinessSelector from '@/components/business-selector/index.vue';
import { timeFormatter } from '@/common/util';
import type { IBizResourcesTicketsParam, IOpResourcesTicketsParam } from '@/typings/resourcePlan';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      required: true,
    },
  },
  emits: ['search'],
  setup(props, { emit }) {
    const { Option } = Select;
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const opList = ref([]);
    const typeList = ref([]);
    const planProductsList = ref([]);
    const statues = ref([]);
    const planProductsLoading = ref(false);
    const statueLoading = ref(false);
    const typeLoading = ref(false);
    const opLoading = ref(false);

    const getDefaultSearchModel = (): IBizResourcesTicketsParam | IOpResourcesTicketsParam => {
      if (props.isBiz) {
        return {
          ticket_ids: [],
          statuses: [],
          ticket_types: [],
          applicants: [],
          submit_time_range: undefined,
        };
      }
      return {
        bk_biz_ids: [],
        op_product_ids: [],
        plan_product_ids: [],
        ticket_ids: [],
        statuses: [],
        ticket_types: [],
        applicants: [],
        submit_time_range: undefined,
      };
    };

    const searchModel = ref<IBizResourcesTicketsParam | IOpResourcesTicketsParam>(
      JSON.parse(JSON.stringify(getDefaultSearchModel())),
    );

    // 获取单据状态列表
    const getStatues = async () => {
      try {
        statueLoading.value = true;
        const res = await resourcePlanStore.getStatusList();
        statues.value = res?.data?.details || [];
      } catch (error: unknown) {
        Message({
          message:
            (
              error as {
                message?: string;
              }
            )?.message || error,
          theme: 'error',
        });
      } finally {
        statueLoading.value = false;
      }
    };

    // 获取 类型列表
    const getTicketTypes = async () => {
      try {
        typeLoading.value = true;
        const res = await resourcePlanStore.getTicketTypesList();
        typeList.value = res?.data?.details || [];
      } catch (error: unknown) {
        Message({
          message:
            (
              error as {
                message?: string;
              }
            )?.message || error,
          theme: 'error',
        });
      } finally {
        typeLoading.value = false;
      }
    };

    const getOpsList = async () => {
      try {
        opLoading.value = true;
        const res = await resourcePlanStore.getOpProductsList();
        opList.value = res?.data?.details || [];
      } catch (error: unknown) {
        Message({
          message:
            (
              error as {
                message?: string;
              }
            )?.message || error,
          theme: 'error',
        });
      } finally {
        opLoading.value = false;
      }
    };

    const getPlanProductsList = async () => {
      try {
        planProductsLoading.value = true;
        const res = await resourcePlanStore.getPlanProductsList();
        planProductsList.value = res?.data?.details || [];
      } catch (error: unknown) {
        Message({
          message:
            (
              error as {
                message?: string;
              }
            )?.message || error,
          theme: 'error',
        });
      } finally {
        planProductsLoading.value = false;
      }
    };

    const handleSearch = () => {
      emit('search', searchModel.value);
    };

    const handleReset = () => {
      searchModel.value = JSON.parse(JSON.stringify(getDefaultSearchModel()));
      emit('search', undefined);
    };

    const handleInputTicket = (val: string) => {
      searchModel.value.ticket_ids = val.split(';').filter((v) => v);
    };

    const handleChangeDate = (val: string[]) => {
      if (val[0] && val[1]) {
        searchModel.value.submit_time_range = {
          start: timeFormatter(val[0], 'YYYY-MM-DD'),
          end: timeFormatter(val[1], 'YYYY-MM-DD'),
        };
      } else {
        searchModel.value.submit_time_range = undefined;
      }
    };

    onBeforeMount(() => {
      getStatues();
      getTicketTypes();
      if (!props.isBiz) {
        getOpsList();
        getPlanProductsList();
      }
    });

    return () => (
      <Panel class={cssModule['mb-16']}>
        <div class={cssModule['search-grid']}>
          {!props.isBiz && (
            <div>
              <div class={cssModule['search-label']}>{t('业务')}</div>
              <BusinessSelector
                v-model={(searchModel.value as IOpResourcesTicketsParam).bk_biz_ids}
                multiple={true}
                authed={true}
                autoSelect={true}
                isShowAll={true}
              />
            </div>
          )}
          {!props.isBiz && (
            <div>
              <div class={cssModule['search-label']}>{t('运营产品')}</div>
              <Select
                multiple
                v-model={(searchModel.value as IOpResourcesTicketsParam).op_product_ids}
                loading={opLoading.value}>
                {opList.value.map((item) => (
                  <Option key={item.op_product_id} id={item.op_product_id} name={item.op_product_name}>
                    {item.op_product_name}
                  </Option>
                ))}
              </Select>
            </div>
          )}
          {!props.isBiz && (
            <div>
              <div class={cssModule['search-label']}>{t('规划产品')}</div>
              <Select
                multiple
                v-model={(searchModel.value as IOpResourcesTicketsParam).plan_product_ids}
                loading={planProductsLoading.value}>
                {planProductsList.value.map((item) => (
                  <Option key={item.plan_product_id} id={item.plan_product_id} name={item.plan_product_name}>
                    {item.plan_product_name}
                  </Option>
                ))}
              </Select>
            </div>
          )}
          <div>
            <div class={cssModule['search-label']}>{t('类型')}</div>
            <Select multiple v-model={searchModel.value.ticket_types} loading={typeLoading.value}>
              {typeList.value.map((item) => (
                <Option key={item.ticket_type} id={item.ticket_type} name={item.ticket_type_name}>
                  {item.ticket_type_name}
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
            <div class={cssModule['search-label']}>{t('单据状态')}</div>
            <Select multiple v-model={searchModel.value.statuses} loading={statueLoading.value}>
              {statues.value.map((item) => (
                <Option key={item.status} name={item.status_name} id={item.status}>
                  {item.status_name}
                </Option>
              ))}
            </Select>
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('提单人')}</div>
            <MemberSelect v-model={searchModel.value.applicants} />
          </div>
          <div>
            <div class={cssModule['search-label']}>{t('提单时间')}</div>
            <DatePicker
              modelValue={[searchModel.value.submit_time_range?.start, searchModel.value.submit_time_range?.end]}
              onChange={(val: string[]) => handleChangeDate(val)}
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
