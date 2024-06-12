import { defineComponent } from 'vue';
import './index.scss';
import useFormModel from '@/hooks/useFormModel';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { Button, Form, Input } from 'bkui-vue';
import BusinessSelector from '@/components/business-selector/index.vue';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import ApplicationStatusSelector from '@/components/scr/application-status-selector';
import ScrDatePicker from '@/components/scr/scr-date-picker';
import MemberSelect from '@/components/MemberSelect';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { useRoute, useRouter } from 'vue-router';
import moment from 'moment';
import WName from '@/components/w-name';
import { HelpDocumentFill } from 'bkui-vue/lib/icon';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';

const { FormItem } = Form;
export default defineComponent({
  setup() {
    const businessMapStore = useBusinessMapStore();
    const { transformApplyStages } = useApplyStages();
    const { formModel, resetForm } = useFormModel({
      bkBizId: businessMapStore.authedBusinessList?.[0]?.id,
      requireType: [],
      stage: [],
      orderId: [],
      dateRange: [],
      user: [],
    });
    const { columns } = useColumns('applicationList');
    const router = useRouter();
    const route = useRoute();
    const { CommonTable, getListData, isLoading } = useTable({
      tableOptions: {
        columns: [
          {
            label: '单号/子单号',
            render: ({ data }: any) => {
              return (
                <div class={'flex-row align-item-center'}>
                  <Button theme='primary' text onClick={() => {
                    router.push({
                      name: 'host-application-detail',
                      params: {
                        id: data.order_id,
                      }
                    });
                  }}>
                    {data.order_id}
                  </Button>
                  <br/>
                  <p class={'ml8 sub-order-txt'}>子单号: {data.suborder_id || '无'}</p>
                </div>
              );
            },
          },
          {
            label: '单据状态',
            field: 'stage',
            width: 250,
            render: ({ data }: any) => {
              const { stage, createAt, modify_time: modifyTime } = data;
              const diffHours = moment(new Date()).diff(moment(createAt), 'hours');
              const isAbnormal = diffHours >= 2 && stage === 'RUNNING';
      
              const stageClass = (stage: string) => {
                if (stage === 'UNCOMMIT') return 'c-text-3';
                if (stage === 'AUDIT') return 'c-text-2';
                if (stage === 'DONE') return 'c-success';
                if (isAbnormal) return 'c-warning';
                if (stage === 'RUNNING') return 'c-text-1';
                if (stage === 'TERMINATE') return 'c-danger';
                if (stage === 'SUSPEND') return 'c-danger';
              };
      
              const abnormalStatus = () => {
                if (stage === 'SUSPEND') {
                  return (
                    <div
                      class={'flex-row align-item-center'}
                      v-bk-tooltips={{
                        content: (
                          <span>
                            {modifyTime < 2 ? (
                              <span>
                                建议
                                <Button size='small' text theme={'primary'} class={'ml8'}>
                                  修改需求重试
                                </Button>
                              </span>
                            ) : (
                              <span>
                                请查看详情后联系 <WName name={'BK助手'} class={'ml8'}></WName> 进行处理
                              </span>
                            )}
                          </span>
                        ),
                      }}>
                      备货状态异常 <HelpDocumentFill fill='#ffbb00' width={12} height={12} class={'ml4'} />
                    </div>
                  );
                }
                return null;
              };
      
              const modifyButton = () => {
                return (
                  <Button size='small' text theme={'primary'} class={'ml8'}>
                    修改需求重试
                  </Button>
                );
              };
      
              const progressButton = () => {
                return (
                  <Button size='small' text theme={'primary'} class={'ml8'}>
                    查看详情
                  </Button>
                );
              };
      
              return (
                <div>
                  <p class={stageClass(stage)}>
                  { stage !== 'SUSPEND'  &&transformApplyStages(stage) }
                    {abnormalStatus()}
                  </p>
                  {stage === 'SUSPEND' && modifyTime < 2 ? modifyButton() : null}
                  {['RUNNING', 'DONE', 'SUSPEND'].includes(stage) ? progressButton() : null}
                </div>
              );
            },
          },
          ...columns,
        ],
        extra: {
          border: ['row', 'col', 'outer'],
        }
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => ({
        url: '/api/v1/woa/task/findmany/apply',
        payload: removeEmptyFields({
          bkBizId: formModel.bkBizId,
          orderId: String(formModel.orderId).split('\n').map(v => +v),
          // suborderId: formModel.suborderId,
          bkUsername: formModel.user,
          stage: formModel.stage,
          start: formModel.dateRange[0],
          end: formModel.dateRange[1],
          // page: formModel.page,
          requireType: formModel.requireType
        })
      })
    });
    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'}>
          <Form model={formModel} class={'scr-form-wrapper'}>
            <FormItem label='业务'>
              <BusinessSelector v-model={formModel.bkBizId} authed />
            </FormItem>
            <FormItem label='需求类型'>
              <RequirementTypeSelector v-model={formModel.requireType} multiple />
            </FormItem>
            <FormItem label='单据状态'>
              <ApplicationStatusSelector v-model={formModel.stage} multiple />
            </FormItem>
            <FormItem label='单号'>
              <Input
                v-model={formModel.orderId}
                type='textarea'
                autosize
                resize={false}
                placeholder='请输入单号,多个换行分割'
              />
            </FormItem>
            <FormItem label='申请时间'>
              <ScrDatePicker v-model={formModel.dateRange} />
            </FormItem>
            <FormItem label='申请人'>
              <MemberSelect v-model={formModel.user} />
            </FormItem>
            <Button theme={'primary'} onClick={() => {
              getListData();
            }} class={'ml24 mr8'} loading={isLoading.value}>
              查询
            </Button>
            <Button onClick={() => {
              resetForm();
              getListData();
            }}>清空</Button>
          </Form>
        </div>
        <Button theme='primary' onClick={() => {
          router.push({
            path: '/ziyanScr/hostApplication/apply',
            query: route.query,
          });
        }} class={'ml24'}>
            新增申请
          </Button>
        <div class={'table-container'}>
          <CommonTable />
        </div>
      </div>
    );
  },
});
