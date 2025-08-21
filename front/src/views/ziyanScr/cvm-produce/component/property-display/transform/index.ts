import { useDiskTypes } from './diskType';
import { useImages } from './image';
import { useBusiness } from '@/views/ziyanScr/host-recycle/field-dictionary/bkBizId';

const { getDiskTypesName } = useDiskTypes();

const { getImageName } = useImages();

const { getBusinessNameById } = useBusiness();

export { getDiskTypesName, getImageName, getBusinessNameById };

export * from './networkType';
