Deployment configurations for additional services. Currently suitable for dev deployments only.

## Running

```shell
docker compose up
```

To run locally use ssh reverse port forwarding. For example:
```shell
ssh -R 80:127.0.0.1:1337 user@domain.ru
```