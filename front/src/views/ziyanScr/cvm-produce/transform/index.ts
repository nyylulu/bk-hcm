import { useCvmProduceStatus } from './cvmProduceStatus';
import { useRequireTypes } from './requireType';

const { statusList, getCvmProduceStatus } = useCvmProduceStatus();
const { getTypeCn } = useRequireTypes();

export { statusList, getCvmProduceStatus, getTypeCn };
