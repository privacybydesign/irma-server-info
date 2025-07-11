Collects (minimal) info on IRMA servers

## Getting started
1. Creaate a `.secrets` directory in the root of the project.
2. Add a `conf.yaml` file in the `.secrets` directory with the following content:
```yaml
Port: 8086
DbHost: localhost
DbUser: serverinfo
DbPass: serverinfo
DbName: serverinfo
```

Run the following command to start the server:
```bash
docker compose up
```
