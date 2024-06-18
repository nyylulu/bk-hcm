import { defineComponent, ref, watch, onMounted } from 'vue';
import { getCapacity } from '@/api/host/cvm';
import { isEqual } from 'lodash';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { Card, Popover } from 'bkui-vue';
export default defineComponent({
  props: {
    params: {
      type: Object,
      default: () => {},
    },
  },
  setup(props) {
    const cvmCapacity = ref([]);
    const fetchCapacityList = async () => {
      const { region, zone, device_type } = props.params;
      if (!(region && zone && device_type)) return;
      const res = await getCapacity(props.params);
      cvmCapacity.value =
        res?.data?.info?.map((item) => {
          return {
            ...item,
            zoneCn: getZoneCn(item.zone),
          };
        }) || [];
    };
    watch(
      () => props.params,
      (newVal, oldVal) => {
        if (!isEqual(newVal, oldVal)) {
          fetchCapacityList();
        }
      },
      { deep: true, immediate: true },
    );
    onMounted(() => {});
    return () => (
      <Card class='cvm-capacity-card' showHeader={false}>
        {{
          default: () => {
            if (!cvmCapacity.value.length) {
              return (
                <div>
                  最大可申请量<span class='cvm-capacity-value'>0</span>
                </div>
              );
            }
            return (
              <div class='cvm-capacity'>
                {cvmCapacity.value.map((item, index) => {
                  return (
                    <>
                      {index < 5 ? (
                        <div class='cvm-capacity-item'>
                          <div>
                            {item.zoneCn}最大可申请量<span class='cvm-capacity-value'>{item.max_num}</span>
                          </div>
                          <Popover theme='light'>
                            {{
                              default: () => <div class='cvm-capacity-detail'>(计算明细)</div>,
                              content: () => (
                                <div>
                                  {item.max_info.map((subItem) => {
                                    return (
                                      <div>
                                        {subItem.key}
                                        <span class='sub-capacity-value'>{subItem.value}</span>
                                      </div>
                                    );
                                  })}
                                </div>
                              ),
                            }}
                          </Popover>
                        </div>
                      ) : null}
                    </>
                  );
                })}
              </div>
            );
          },
        }}
      </Card>
    );
  },
});
