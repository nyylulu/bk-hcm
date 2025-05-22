export default (cloudAreaMap: Map<number, any>) => [...cloudAreaMap].filter(([key]) => key !== 0);
