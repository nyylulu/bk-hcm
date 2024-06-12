import { defineComponent, ref, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import cssModule from './index.module.scss';

import type { IPlanTicket } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    modelValue: Object as PropType<IPlanTicket>,
  },

  emits: ['update:modelValue'],

  setup(props, { emit, expose }) {
    const { t } = useI18n();

    const placeholder = `输入该预测需求单的详细说明，如使用背景，使用需求场景等，字数不少于20字，不超过1024字
主要包括内容如下： 
一、需求概述：**业务，因***原因，新增资源总量**台（**核心CPU）；
二、资源详细说明：
（1）建设容量：当前单OS承载**人/核，新增***核算，新增资源可支持**建设容量（可支持的最大同时在线用户，按CPU约70%评估）；
（2）主要模块资源需求说明，如游戏核心模块、DB模块等；
（3）游戏玩法说明及主要资源需求模块功能说明
三、资源退回计划。

新业务上线，详细模板：
一、背景：
   需求背景表述*****

二、业务资源情况：
1、总建设容量****；
2、游戏大厅模块，资源量**核心（*C*G)
3、DB模块，资源量**核算（IT5.8Xlarge128)
4、资源主要在游戏对局模块
（1）游戏初期：单核资源预计承**人，每局**真人，***AI，单台OS（**C）承载***人，预计最大量有***%的用户会在新手局，约***W用户，需要资源量*****台
（2）游戏运营期：单核资源预计承**人，，每局***真人(理论值)，最高单OS承载能力***人；业务后期，视在线情况调整服务器数量。
（3）针对资源使用（游戏初期**真人***AI）数据压测，单服务器分配***场对局，CPU利用率约***%左右；
（4）剧情玩法说明：****

需求资源合计：***核心；

三、退还计划
1、预计业务上线**周后，可以退还暖局机器***核心；
2、预计业务上线**周后，平均真人到**人，可以退还资源**核心(两次合计**w)。
3、后续据游戏在线情况，调整资源投入量。`;
    const rules = {
      remark: [
        {
          validator: (value: string) => value.length > 20,
          message: t('字数不少于20字'),
          trigger: 'change',
        },
      ],
    };
    const fromRef = ref();

    const updateModelValue = (value: string) => {
      emit('update:modelValue', {
        ...props.modelValue,
        remark: value,
      });
    };

    const validate = () => {
      return fromRef.value.validate();
    };

    expose({
      validate,
    });

    return () => (
      <Panel title={t('预测信息')}>
        <bk-form form-type='vertical' ref={fromRef} rules={rules} model={props.modelValue} class={cssModule.home}>
          <bk-form-item label={t('预测说明')} property='remark' required>
            <bk-input
              type='textarea'
              clearable
              rows={10}
              maxlength={1200}
              placeholder={placeholder}
              modelValue={props.modelValue.remark}
              onChange={updateModelValue}
            />
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
