FROM migrate/migrate:latest
WORKDIR /migrations
COPY shared/migrations /migrations

ENTRYPOINT [ "migrate" ]