import { defineComponent, ref, computed, watch, onMounted, useTemplateRef } from 'vue';
import { merge, cloneDeep, isEqual } from 'lodash';
import { getImages, getDeviceTypesDetails } from '@/api/host/cvm';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import DiskTypeSelect from '@/views/ziyanScr/hostApplication/components/DiskTypeSelect';
import ImageDialog from './image-dialog';
import CvmMaxCapacity from '@/views/ziyanScr/components/cvm-max-capacity/index.vue';
import { Alert, Checkbox, Form, Input, Select, Radio, Tag } from 'bkui-vue';
import { HelpFill } from 'bkui-vue/lib/icon';
import CvmVpcSelector from '@/views/ziyanScr/components/cvm-vpc-selector/index.vue';
import CvmSubnetSelector from '@/views/ziyanScr/components/cvm-subnet-selector/index.vue';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import InheritPackageFormItem, {
  type RollingServerHost,
} from '@/views/ziyanScr/rolling-server/inherit-package-form-item/index.vue';
import PlanLinkAlert from '@/views/ziyanScr/hostApplication/plan/plan-link-alert.vue';
import type { CvmDeviceType, DeviceType } from '@/views/ziyanScr/components/devicetype-selector/types';
import usePlanDeviceType from '@/views/ziyanScr/hostApplication/plan/usePlanDeviceType';
import './index.scss';

const { FormItem } = Form;

export default defineComponent({
  components: {
    AreaSelector,
    ZoneSelector,
    DiskTypeSelect,
    ImageDialog,
  },
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '创建单据',
    },
    requireType: {
      type: Number,
      default: 1,
    },
    resourceType: {
      type: String,
      default: 'QCLOUDCVM',
    },
    dataInfo: {
      type: Object,
      default: () => {
        return {};
      },
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const { cvmChargeTypes, cvmChargeTypeNames, cvmChargeTypeTips, getMonthName } = useCvmChargeType();

    const modelForm = ref({
      replicas: 1,
      antiAffinityLevel: 'ANTI_NONE',
      remark: '',
      enableDiskCheck: false,
      spec: {
        region: '',
        zone: '',
        device_type: '',
        image_id: '',
        disk_size: 0,
        disk_type: 'CLOUD_PREMIUM',
        networkType: 'TENTHOUSAND', // 写成一个常量
        vpc: '',
        subnet: '',
        charge_type: cvmChargeTypes.PREPAID,
        charge_months: 36,
        inherit_instance_id: '',
      },
      bk_asset_id: '',
    });
    const advancedSettingVisible = ref(false);
    const rulesForm = {
      'spec.device_type': [{ required: true, message: '请选择机型', trigger: 'change' }],
      'spec.region': [{ required: true, message: '请选择地域', trigger: 'change' }],
      'spec.zone': [{ required: true, message: '请选择园区', trigger: 'change' }],
      replicas: [{ required: true, message: '请选择需求数量', trigger: 'change' }],
      'spec.imageId': [{ required: true, message: '请选择镜像', trigger: 'change' }],
      'spec.subnet': [
        {
          required: true,
          validator: (value) => {
            if (modelForm.value.spec.vpc && !value) {
              advancedSettingVisible.value = true;
              return false;
            }
            return true;
          },
          message: '选择 VPC 后必须选择子网',
          trigger: 'submit',
        },
      ],
    };
    const options = ref({
      regions: [],
      zones: [],
      images: [],
      vpcs: [],
      subnets: [],
      deviceTypes: [],
    });
    watch(
      modelForm,
      () => {
        modelForm.value.spec.charge_months =
          modelForm.value.spec.charge_type === cvmChargeTypes.PREPAID
            ? modelForm.value.spec.charge_months ?? 36
            : undefined;
        emit('update:modelValue', cloneDeep(modelForm.value));
      },
      { deep: true },
    );
    const isHasImgDialog = ref(false);
    watch(
      () => modelForm.value.spec.image_id,
      (newVal) => {
        // 图像弹框的显示
        if (['img-bh86p0sv', 'img-r5igp4bv'].includes(newVal)) {
          isHasImgDialog.value = true;
        }
      },
    );
    const showDiskCheck = ref(false);

    const cvmCapacityParams = computed(() => {
      return {
        region: modelForm.value.spec.region,
        zone: modelForm.value.spec.zone,
        vpc: modelForm.value.spec.vpc,
        device_type: modelForm.value.spec.device_type,
        subnet: modelForm.value.spec.subnet,
        charge_type: modelForm.value.spec.charge_type,
        require_type: props.requireType,
      };
    });
    const initModel = () => {
      merge(modelForm.value, cloneDeep(props.modelValue));
    };
    const clearRegionRelationItems = () => {
      modelForm.value.spec.zone = '';
      modelForm.value.spec.image_id = '';
      modelForm.value.spec.device_type = '';
    };
    const loadImages = () => {
      getImages({
        region: [modelForm.value.spec.region],
      })
        .then((res) => {
          options.value.images = res?.data?.info || [];
          if (!modelForm.value.spec.image_id) {
            modelForm.value.spec.image_id = 'img-fjxtfi0n';
          }
        })
        .catch(() => {
          options.value.images = [];
        });
    };
    const loadRegionRelationOpts = () => {
      loadImages();
    };
    const handleRegionChange = () => {
      clearRegionRelationItems();
      loadRegionRelationOpts();
    };
    const clearZoneRelationItems = () => {
      modelForm.value.spec.vpc = '';
      modelForm.value.spec.subnet = '';
      modelForm.value.spec.device_type = '';
    };

    const handleZoneChange = () => {
      clearZoneRelationItems();
    };
    const loadVpcRelationOpts = () => {
      modelForm.value.spec.subnet = '';
    };
    const loadDeviceTypeDetail = () => {
      const rules = [];
      ['region', 'zone', 'device_type'].map((item) => {
        if (modelForm.value.spec[item]) {
          rules.push({
            field: item,
            operator: 'equal',
            value: modelForm.value.spec[item],
          });
        }
        return null;
      });
      let params = {};
      if (rules.length) {
        params = {
          filter: {
            condition: 'AND',
            rules,
          },
        };
      }
      getDeviceTypesDetails(params).then((res) => {
        const list = ['高IO型', '大数据型'];
        if (list.includes(res.data.info[0]?.device_group)) {
          showDiskCheck.value = true;
        } else {
          modelForm.value.enableDiskCheck = false;
          showDiskCheck.value = false;
        }
      });
    };
    watch(
      () => props.modelValue,
      (newVal) => {
        if (isEqual(newVal, modelForm.value)) return;
        initModel();
        loadRegionRelationOpts();
        loadVpcRelationOpts();
      },
      { immediate: true, deep: true },
    );
    watch(
      () => modelForm.value.spec.device_type,
      () => {
        loadDeviceTypeDetail();
      },
    );
    const isShowAdvanceSet = ref(false);
    const changeAdvanceItem = () => {
      isShowAdvanceSet.value = !isShowAdvanceSet.value;
    };
    const modelFormRef = ref(null);
    const validate = () => {
      return modelFormRef.value.validate();
    };
    const clearValidate = () => {
      modelFormRef.value.clearValidate();
    };

    const cvmDevicetypeParams = computed(() => {
      const { region, zone } = modelForm.value.spec;
      return {
        region,
        zone: zone !== 'cvm_separate_campus' ? zone : undefined,
      };
    });

    const isRollingServer = computed(() => props.requireType === 6);
    const isSpecialRequirement = computed(() => [6, 7].includes(props.requireType));
    watch(isSpecialRequirement, (val) => {
      if (!val) {
        modelForm.value.bk_asset_id = '';
        modelForm.value.spec.inherit_instance_id = '';
        return;
      }
    });

    // 滚服继承套餐的机器
    let rollingServerHost: RollingServerHost = null;

    // 机型排序
    const deviceTypeSortFn = (a: DeviceType, b: DeviceType) => {
      // 非滚服、非小额绿通，走预测
      if (!isSpecialRequirement.value) {
        const set = computedAvailableDeviceTypeSet.value;
        return Number(set.has(b.device_type)) - Number(set.has(a.device_type));
      }
      // 滚服、小额绿通
      const aDeviceTypeClass = (a as CvmDeviceType).device_type_class;
      const bDeviceTypeClass = (b as CvmDeviceType).device_type_class;
      if (aDeviceTypeClass === 'CommonType' && bDeviceTypeClass === 'SpecialType') return -1;
      if (aDeviceTypeClass === 'SpecialType' && bDeviceTypeClass === 'CommonType') return 1;
      return 0;
    };
    // 机型选项禁用
    const deviceTypeOptionDisabledCallback = (option: DeviceType) => {
      // 非滚服、非小额绿通
      if (!isSpecialRequirement.value) {
        return !computedAvailableDeviceTypeSet.value.has(option.device_type);
      }
      // 滚服、小额绿通
      const { device_type_class, device_group } = option as CvmDeviceType;
      return (
        'SpecialType' === device_type_class ||
        (isRollingServer.value && device_group !== rollingServerHost?.device_group)
      );
    };
    // 机型选项禁用tips
    const deviceTypeOptionDisabledTipsCallback = (option: DeviceType) => {
      // 非滚服、非小额绿通
      if (!isSpecialRequirement.value) return '当前机型不在有效预测范围内';
      // 滚服、小额绿通
      const { device_type_class, device_group } = option as CvmDeviceType;
      if (device_type_class === 'SpecialType') return '专用机型不允许选择';
      if (isRollingServer.value && device_group !== rollingServerHost?.device_group) return '机型族不匹配';
    };

    // 预测-机型
    const deviceTypeSelectorRef = useTemplateRef<typeof DevicetypeSelector>('device-type-selector');
    const selectedChargeType = computed(() => modelForm.value.spec.charge_type);
    const {
      isPlanedDeviceTypeLoading,
      availableDeviceTypeSet,
      computedAvailableDeviceTypeSet,
      hasPlanedDeviceType,
      getPlanedDeviceType,
    } = usePlanDeviceType(deviceTypeSelectorRef, selectedChargeType);
    // 获取有效预测范围内的机型
    watch(
      [() => props.requireType, () => modelForm.value.spec.region, () => modelForm.value.spec.zone],
      async ([require_type, region, zone]) => {
        if (!require_type || !region || !zone || props.resourceType !== 'QCLOUDCVM' || isSpecialRequirement.value)
          return;

        // 业务为：资源运营服务
        await getPlanedDeviceType(931, require_type, region, zone);

        if (availableDeviceTypeSet.prepaid.size === 0) {
          modelForm.value.spec.charge_type = cvmChargeTypes.POSTPAID_BY_HOUR;
        }
      },
      { deep: true },
    );

    const isSubmitDisabled = computed(() => !isSpecialRequirement.value && !hasPlanedDeviceType.value);

    expose({ isSubmitDisabled, validate, clearValidate });

    onMounted(() => {
      initModel();
      loadRegionRelationOpts();
      loadVpcRelationOpts();
    });

    return () => (
      <>
        <Form ref={modelFormRef} model={modelForm.value} rules={rulesForm} class='cvm-produce-create-form'>
          <div class='flex-row'>
            <FormItem label='地域' required property='spec.region'>
              <area-selector
                v-model={modelForm.value.spec.region}
                params={{ resourceType: props.resourceType }}
                onChange={handleRegionChange}
                class='i-form-control'
              />
            </FormItem>

            <FormItem label='园区' required property='spec.zone'>
              <zone-selector
                v-model={modelForm.value.spec.zone}
                params={{ resourceType: props.resourceType, region: modelForm.value.spec.region }}
                disabled={!modelForm.value.spec.region}
                placeholder={!modelForm.value.spec.region ? '请先选择地域' : '请选择园区'}
                onChange={handleZoneChange}
                class='i-form-control'
              />
            </FormItem>
          </div>

          {/* 预测指引 */}
          {!isSpecialRequirement.value &&
            modelForm.value.spec.zone &&
            !hasPlanedDeviceType.value &&
            !isPlanedDeviceTypeLoading.value && (
              <FormItem style='margin-top: -12px'>
                <PlanLinkAlert bkBizId={931} />
              </FormItem>
            )}

          <FormItem label='机型' required property='spec.device_type'>
            <DevicetypeSelector
              ref='device-type-selector'
              v-model={modelForm.value.spec.device_type}
              resourceType='cvm'
              params={cvmDevicetypeParams.value}
              disabled={!modelForm.value.spec.zone}
              isLoading={isPlanedDeviceTypeLoading.value}
              placeholder={!modelForm.value.spec.zone ? '请先选择园区' : '请选择机型'}
              sort={deviceTypeSortFn}
              optionDisabled={deviceTypeOptionDisabledCallback}
              optionDisabledTipsContent={deviceTypeOptionDisabledTipsCallback}
              class='i-form-control'>
              {{
                option: (option: CvmDeviceType) => {
                  const { device_type, device_type_class, device_group } = option;
                  const isSpecialType = device_type_class === 'SpecialType';
                  return (
                    <>
                      <span>{device_type}</span>
                      <Tag class='ml12' theme={isSpecialType ? 'danger' : 'success'} size='small'>
                        {isSpecialType ? '专用机型' : '通用机型'}
                      </Tag>
                      {device_group && (
                        <Tag class='ml12' size='small'>
                          {device_group}
                        </Tag>
                      )}
                    </>
                  );
                },
              }}
            </DevicetypeSelector>
            {showDiskCheck.value && (
              <Checkbox v-model={modelForm.value.enableDiskCheck}>
                <div class='flex-row align-items-center'>
                  <span>本地盘压测</span>
                  <HelpFill
                    class='ml8'
                    v-bk-tooltips={{ content: '本地盘压测耗时较长，尤其大数据类设备测试耗时超过1小时，请注意！' }}
                  />
                </div>
              </Checkbox>
            )}
          </FormItem>

          <FormItem label='镜像' required property='spec.image_id'>
            <Select
              v-model={modelForm.value.spec.image_id}
              clearable
              disabled={!modelForm.value.spec.region}
              placeholder={!modelForm.value.spec.region ? '请先选择地域' : '请选择镜像'}
              class='i-form-control'>
              {options.value.images.map(({ image_id, image_name }) => {
                return <Select.Option key={image_id} name={image_name} id={image_id} />;
              })}
            </Select>
          </FormItem>

          <FormItem label='数据盘' class='i-disk-form-item'>
            <div class='i-content '>
              <div class='i-form-control flex-row'>
                <disk-type-select v-model={modelForm.value.spec.disk_type} class='i-disk-type-select' />
                <Input v-model={modelForm.value.spec.disk_size} type='number' min={0} step={10} max={16000} />
              </div>
              <span class='i-unit'>G</span>
              <HelpFill v-bk-tooltips={{ content: '最大为 16T(16000 G)，且必须为 10 的倍数' }} />
            </div>
            {modelForm.value.spec.disk_type === 'CLOUD_SSD' && (
              <div class='i-tips'>SSD 云硬盘的运营成本约为高性能云盘的 4 倍，请合理评估使用。</div>
            )}
          </FormItem>

          <FormItem label='需求数量' required property='replicas'>
            <div class='flex-row align-items-center'>
              <Input v-model={modelForm.value.replicas} type='number' min={1} max={1000} class='i-form-control' />
              <HelpFill
                class='ml8'
                v-bk-tooltips={{ content: '如果需求数量超过最大可申请量，请提单后联系管理员 forestchen, dommyzhang' }}
              />
            </div>
            <CvmMaxCapacity params={cvmCapacityParams.value} />
          </FormItem>

          {/* 滚服项目-继承套餐 */}
          {isRollingServer.value && (
            <InheritPackageFormItem
              v-model={modelForm.value.bk_asset_id}
              region={modelForm.value.spec.region}
              onValidateSuccess={(host) => {
                const { instance_charge_type: chargeType, charge_months: chargeMonths } = host;
                modelForm.value.spec.charge_type = chargeType;
                modelForm.value.spec.charge_months = chargeType === cvmChargeTypes.PREPAID ? chargeMonths : undefined;
                modelForm.value.spec.inherit_instance_id = host.bk_cloud_inst_id;
                // 机型族与上次数据不一致时需要清除机型选择
                if (
                  rollingServerHost &&
                  host.device_group !== rollingServerHost.device_group &&
                  modelForm.value.spec.device_type
                ) {
                  modelForm.value.spec.device_type = '';
                }
                rollingServerHost = host;
              }}
              onValidateFailed={() => {
                // 恢复默认值
                modelForm.value.spec.charge_type = cvmChargeTypes.PREPAID;
                modelForm.value.spec.charge_months = 36;
                if (modelForm.value.spec.device_type) {
                  modelForm.value.spec.device_type = '';
                }
                rollingServerHost = null;
              }}
            />
          )}

          <FormItem label='计费模式' required property='spec.charge_type'>
            {isSpecialRequirement.value ? (
              // 滚服项目、小额绿通
              <Radio.Group
                v-model={modelForm.value.spec.charge_type}
                type='card'
                disabled={isRollingServer.value}
                v-bk-tooltips={{ content: '继承原有套餐，计费模式不可选', disabled: !isRollingServer.value }}
                class='i-form-control'>
                <Radio.Button label={cvmChargeTypes.PREPAID}>{cvmChargeTypeNames[cvmChargeTypes.PREPAID]}</Radio.Button>
                <Radio.Button label={cvmChargeTypes.POSTPAID_BY_HOUR}>
                  {cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR]}
                </Radio.Button>
              </Radio.Group>
            ) : (
              // 其他需求类型（走预测）
              <Radio.Group
                v-model={modelForm.value.spec.charge_type}
                type='card'
                disabled={isRollingServer.value}
                v-bk-tooltips={{ content: '继承原有套餐，计费模式不可选', disabled: !isRollingServer.value }}
                class='i-form-control'>
                <Radio.Button
                  label={cvmChargeTypes.PREPAID}
                  disabled={availableDeviceTypeSet.prepaid.size === 0}
                  v-bk-tooltips={{
                    content: '当前地域无有效的预测需求，请提预测单后再按量申请',
                    disabled: !modelForm.value.spec.zone || availableDeviceTypeSet.prepaid.size > 0,
                  }}>
                  {cvmChargeTypeNames[cvmChargeTypes.PREPAID]}
                </Radio.Button>
                <Radio.Button
                  label={cvmChargeTypes.POSTPAID_BY_HOUR}
                  disabled={availableDeviceTypeSet.postpaid.size === 0}
                  v-bk-tooltips={{
                    content: '当前地域无有效的预测需求，请提预测单后再按量申请',
                    disabled: !modelForm.value.spec.zone || availableDeviceTypeSet.postpaid.size > 0,
                  }}>
                  {cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR]}
                </Radio.Button>
              </Radio.Group>
            )}
            <Alert theme='info' class='form-item-tips'>
              {{
                title: () => (
                  <>
                    {cvmChargeTypeTips[modelForm.value.spec.charge_type]}
                    <bk-link href='https://crp.woa.com/crp-outside/yunti/news/20' theme='primary' target='_blank'>
                      计费模式说明
                    </bk-link>
                  </>
                ),
              }}
            </Alert>
          </FormItem>

          {modelForm.value.spec.charge_type === cvmChargeTypes.PREPAID && (
            <FormItem label='购买时长' required property='spec.charge_months'>
              <Select
                v-model={modelForm.value.spec.charge_months}
                filterable={false}
                clearable={false}
                disabled={isRollingServer.value}
                v-bk-tooltips={{
                  content: '继承原有套餐包年包月时长，此处的购买时长为剩余时长',
                  disabled: !isRollingServer.value,
                }}
                class='i-form-control'>
                {(function () {
                  const options = isRollingServer.value
                    ? Array.from({ length: 48 }, (v, i) => i + 1)
                    : [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48];
                  return options.map((option) => <bk-option key={option} value={option} name={getMonthName(option)} />);
                })()}
              </Select>
            </FormItem>
          )}

          <FormItem label='备注' property='remark'>
            <Input
              v-model={modelForm.value.remark}
              placeholder='请输入备注'
              show-word-limit
              type='textarea'
              resize={false}
              maxlength={128}
              class='i-form-control'
            />
          </FormItem>

          <div class='advance-set' onClick={changeAdvanceItem}>
            高级设置
          </div>
          {isShowAdvanceSet.value && (
            <FormItem
              label='网络'
              required
              property='spec.subnet'
              v-bk-tooltips={{
                disabled: modelForm.value.spec.zone !== 'cvm_separate_campus',
                content: '园区分Campus 时无法指定子网',
              }}>
              <div class='flex-row'>
                <CvmVpcSelector
                  v-model={modelForm.value.spec.vpc}
                  region={modelForm.value.spec.region}
                  onChange={loadVpcRelationOpts}
                  disabled={modelForm.value.spec.zone === 'cvm_separate_campus' || !modelForm.value.spec.region}
                  class='i-form-control'
                />
                <CvmSubnetSelector
                  v-model={modelForm.value.spec.subnet}
                  region={modelForm.value.spec.region}
                  zone={modelForm.value.spec.zone}
                  vpc={modelForm.value.spec.vpc}
                  disabled={modelForm.value.spec.zone === 'cvm_separate_campus' || !modelForm.value.spec.vpc}
                  class='i-form-control'
                />
              </div>
              <Alert
                class='mt12'
                title='一般需求不需要指定 VPC 和子网，如为 BCS、ODP 等 TKE 场景母机，请提前与平台支持方确认 VPC、子网信息。'
              />
            </FormItem>
          )}
        </Form>
        <image-dialog v-model={isHasImgDialog.value} />
      </>
    );
  },
});
