FROM hub.bktencent.com/blueking/bk-hcm-base:v1.1

# RUN mkdir -p /data/hcm/plugin
COPY bk-hcm-woaserver /data/hcm/
COPY plugin/algorithm /data/hcm/plugin/
CMD ["/data/hcm/bk-hcm-woaserver", "--config-file", "/data/hcm/etc/config.yaml"]
