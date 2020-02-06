version: '{{ .version }}'

{{- if .network_enabled }}
neworks:
  {{ .network_name }}:
    driver: overlay
    external: false
    attachable: true
    ipam:
      config:
        - subnet: {{ .network_subnet }}
{{- end }}

services:

  mariadb:
    image: bitnami/mariadb:{{ .mariadb_version }}
    environment:
      - "MARIADB_PORT_NUMBER=3306"
      - "MARIADB_ROOT_USER=root"
    user: "999"
    {{- if .mariadb_volume_enabled }}
    volumes:
      - "/etc/localtime:/etc/localtime:ro"
    {{- end }}
    deploy:
      mode: global
      placement:
        constraints: [node.platform.os == linux]
      restart_policy:
        condition: on-failure
        delay: 5s
      resources:
        limits:
          cpus: "2.0"
          memory: 2000MB
      update_config:
        parallelism: 1
        delay: 10m
