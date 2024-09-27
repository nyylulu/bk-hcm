import { defineComponent, PropType } from 'vue';
import './index.scss';
import Panel from '@/components/panel';
import { Form } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { useI18n } from 'vue-i18n';
import { AdjustType } from '@/typings/plan';
const { FormItem } = Form;

export default defineComponent({
  props: {
    modelValue: String as PropType<AdjustType>,
    type: String as PropType<AdjustType>,
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const { t } = useI18n();

    const triggerUpdate = (val: string) => {
      emit('update:modelValue', val);
    };

    return () => (
      <Panel class={'mb16'} title={`${'调整类型'}`}>
        <Form formType='vertical'>
          <FormItem label={t('调整方式')}>
            <BkRadioGroup modelValue={props.modelValue} onChange={triggerUpdate}>
              <BkRadioButton
                label={AdjustType.config}
                disabled={props.type === AdjustType.time}
                v-bk-tooltips={{
                  content: t('已延期，不支持调整'),
                  disabled: props.type !== AdjustType.time,
                }}>
                {t('调整配置')}
              </BkRadioButton>
              <BkRadioButton
                label={AdjustType.time}
                disabled={props.type === AdjustType.config}
                v-bk-tooltips={{
                  content: t('已修改配置，不支持调整'),
                  disabled: props.type !== AdjustType.config,
                }}>
                {t('调整时间')}
              </BkRadioButton>
            </BkRadioGroup>
          </FormItem>
        </Form>
      </Panel>
    );
  },
});
