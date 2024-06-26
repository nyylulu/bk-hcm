import { defineComponent, ref, watch, onMounted } from 'vue';
import { Sideslider, Table } from 'bkui-vue';
import { modifyRecord } from '@/api/host/task';
import { useFieldVal } from '@/components/property-list/field-map';
import { dateTimeTransform } from '@/views/ziyanScr/host-recycle/field-dictionary';
export default defineComponent({
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '变更记录',
    },
    showObj: {
      type: Object,
      default: () => ({}),
    },
  },
  emits: ['update:modelValue'],
  setup(props, { attrs, emit }) {
    const { getFieldCn, getFieldCnVal } = useFieldVal();
    const isDisplay = ref(false);
    watch(
      () => props.modelValue,
      (val) => {
        isDisplay.value = val;
        if (val) {
          fetchRecord();
        }
      },
      {
        immediate: true,
      },
    );
    const updateShowValue = () => {
      emit('update:modelValue', false);
    };
    const recordList = ref([]);
    const recordColumns = [
      {
        type: 'expand',
      },
      {
        label: 'ID',
        field: 'id',
      },
      {
        label: '操作人',
        field: 'bk_username',
      },
      {
        label: '操作时间',
        field: 'create_at',
        render: ({ row }: any) => {
          return <span>{dateTimeTransform(row.create_at)}</span>;
        },
      },
    ];
    const expandRowTable = [
      {
        label: '修改属性',
        field: 'key',
        align: 'left',
        render: ({ row }: any) => {
          return <span>{getFieldCn(row.key)}</span>;
        },
      },
      {
        label: '变更前',
        field: 'prev',
        align: 'left',
        render: ({ row }: any) => {
          return <span>{getFieldCnVal(row.key, row.prev)}</span>;
        },
      },
      {
        label: '变更后',
        field: 'cur',
        align: 'left',
        render: ({ row }: any) => {
          return <span>{getFieldCnVal(row.key, row.cur)}</span>;
        },
      },
    ];
    const handleDetail = (detail: any) => {
      return Object.keys(detail?.cur_data).reduce((prev, cur) => {
        const obj = {
          key: cur,
          prev: detail.pre_data[cur],
          cur: detail.cur_data[cur],
          hasChange: detail.pre_data[cur] !== detail.cur_data[cur],
        };
        prev.push(obj);
        return prev;
      }, []);
    };
    const fetchRecord = () => {
      modifyRecord({ suborder_id: [props.showObj.suborderId] })
        .then((res) => {
          const list = res.data?.info || [];
          list.forEach((item) => {
            item.detailList = handleDetail(item.details);
          });
          recordList.value = list;
        })
        .finally(() => {});
    };
    onMounted(() => {});
    return () => (
      <Sideslider
        v-bind={attrs}
        width='1000'
        v-model:isShow={isDisplay.value}
        title={props.title}
        before-close={updateShowValue}>
        {{
          default: () => (
            <div class='sideslider-content'>
              <div class='des-top'>
                <div>
                  <label>所属业务</label>
                  <span>{props.showObj.bkBizId}</span>
                </div>
                <div>
                  <label>变更对象</label>
                  <span>资源申请子单</span>
                </div>
                <div>
                  <label>变更对象ID</label>
                  <span>{props.showObj.suborderId}</span>
                </div>
              </div>
              <Table data={recordList.value} columns={recordColumns}>
                {{
                  expandRow: (row) => (
                    <div class='record-expand'>
                      <Table data={row.detailList} columns={expandRowTable}></Table>
                    </div>
                  ),
                }}
              </Table>
            </div>
          ),
        }}
      </Sideslider>
    );
  },
});
