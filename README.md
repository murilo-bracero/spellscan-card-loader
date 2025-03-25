# spellscan-card-loader
Job that loads new batches of cards into database

## Building

### Binary

Run `make build`, the binary will be built in the `build` folder on the root of the repository.

### Docker

Run `make dockerBuild`, the image will be available as `ghcr.io/murilo-bracero/spellscan-card-loader:latest`

## Running

The following environment variables are required:

- DB_DSN: DSN (URL) of the postgres database, following the pattern: `postgres://{USER}:{PASSWORD}@{HOST}:{PORT}/{DATABASE_NAME}{...ARGS}`
- MEILI_API_KEY: API Key of the Meilisearch instance
- MEILI_URL: URL of the Meilisearch instance
- DB_MAX_CONNECTIONS: Max number of database connections that the job can use

Optional, but highly recommended, environment variables:

- SKIP_DOWNLOAD: If set to true, will use the previously downloaded batch from Scryfall.
- USE_RELEASE_DATE_REFERENCE: If set to true, will get the latest card added to the database and use it as a reference to ignore cards added before it.

### Binary

Run `make run`, the binary will be built in the `build` folder on the root of the repository.

### Docker

Run `make dockerRun`, the image should be available as `ghcr.io/murilo-bracero/spellscan-card-loader:latest`.

### Docker Compose

To use docker compose, create a file named `.env.docker` with the environment variables, and run `docker-compose up`.

It will spin up a postgres container, a meilisearch container, a liquibase container to initialize the database schema and the spellscan-card-loader container.

## License

This project is released under the Mozilla Public License 2.0.