FROM postgres:latest
EXPOSE 5432
WORKDIR /scripts
COPY db/scripts/create_database.sql /scripts/db.sql
COPY db/scripts/create_schema.sql /scripts/schema.sql
CMD psql-f /scripts/db.sql
CMD psql -f /scripts/schema.sql

