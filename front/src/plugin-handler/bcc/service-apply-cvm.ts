import { useAccountSelectorCard } from '@/hooks/useAccountSelectorCard';
import { PluginHandlerType } from '../service-apply-cvm';
import applicationForm from '@/views/ziyanScr/hostApplication/components/application-form';

export const pluginHandler: PluginHandlerType = {
  useAccountSelector: useAccountSelectorCard,
  ApplicationForm: applicationForm,
};
