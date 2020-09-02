FROM postgres:latest
EXPOSE 54321
WORKDIR /scripts
COPY db/scripts/00_create_database.sql /scripts/db.sql
COPY db/scripts/01_create_schema.sql /scripts/schema.sql
COPY db/scripts/02_create_additional_trigger_check.sql /scripts/triger.sql
CMD psql -f /scripts/db.sql
CMD psql -f /scripts/schema.sql

