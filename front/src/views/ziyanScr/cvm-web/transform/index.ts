import { useRegions } from './region';
import { useZones } from './zone';

const { getRegionCn } = useRegions();
const { getZoneCn } = useZones();

export { getRegionCn, getZoneCn };
