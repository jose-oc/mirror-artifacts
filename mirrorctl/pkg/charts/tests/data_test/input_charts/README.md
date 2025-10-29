These are helm charts pulled from their original source. 
They are used to test this package. 
They were pulled with the following commands.

```shell
helm pull grafana/grafana-agent-operator --version 0.5.1 --untar
helm pull grafana/loki --version 5.5.2 --untar
helm pull oci://registry-1.docker.io/bitnamicharts/mariadb --version 12.2.4 --untar
helm pull influxdata/telegraf --version 1.8.28 --untar
```
