import { defineComponent, ref, watch } from 'vue';
import { Button, Dialog } from 'bkui-vue';
export default defineComponent({
  props: {
    modelValue: {
      type: Boolean,
      default: '',
    },
    title: {
      type: String,
      default: '',
    },
  },
  emits: ['update:modelValue'],
  setup(props, { attrs, emit }) {
    const isShow = ref(false);
    watch(
      () => props.modelValue,
      (val) => {
        isShow.value = val;
      },
      { immediate: true },
    );
    const updateSelectedValue = () => {
      emit('update:modelValue', false);
    };
    return () => (
      <Dialog v-bind={attrs} v-model:is-show={isShow.value} title='批量更新'>
        {{
          default: () => (
            <div>
              <div>接公司内核支持团队通知，该操作系统对应内核版本已停止更新，不建议继续使用，具体请参考：</div>
              <div>
                内核的生命周期管理策略
                <a target='_blank' class='link' href='https://iwiki.woa.com/p/1042997715'>
                  https://iwiki.woa.com/p/1042997715
                </a>
              </div>
              <div>
                1.2 + 2.2（final)版本的风险:
                存在内核无法识别出新网段为内网网段的情况（即判断新网段ip为外网）、可能会影响基于时间戳的相关统计功能异常，或存在由timewait带来的性能影响
                , 参考文档
                <a target='_blank' class='link' href='https://iwiki.woa.com/p/4007500196'>
                  https://iwiki.woa.com/p/4007500196
                </a>
                &nbsp;&nbsp;&nbsp;&nbsp;
                <a target='_blank' class='link' href='https://iwiki.woa.com/p/1985132973'>
                  https://iwiki.woa.com/p/1985132973
                </a>
              </div>
              <div>
                当前推荐版本的推广使用情况
                <a target='_blank' class='link' href='https://iwiki.woa.com/p/1985132986'>
                  https://iwiki.woa.com/p/1985132986
                </a>
              </div>

              <div>
                若业务因特殊原因，无法短期内完成切换的，请参考文档走邮件申请
                <a target='_blank' class='link' href='https://iwiki.woa.com/p/4006721481'>
                  https://iwiki.woa.com/p/4006721481
                </a>
              </div>
            </div>
          ),
          footer: () => (
            <Button theme='primary' onClick={updateSelectedValue}>
              确 定
            </Button>
          ),
        }}
      </Dialog>
    );
  },
});
