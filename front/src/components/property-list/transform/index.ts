import { useDiskTypes } from './diskType';
import { useImages } from './image';

import { useBusiness } from '@/views/ziyanScr/host-recycle/field-dictionary/bkBizId';
const { getDiskTypesName } = useDiskTypes();
const { getImageName } = useImages();
export { getDiskTypesName, getImageName };
export const { getBusinessNameById } = useBusiness();

export * from './networkType';
