# Use official PostgreSQL image
FROM postgres:15-alpine

# Copy configuration
COPY postgresql.conf /etc/postgresql/postgresql.conf

# Copy initialization scripts
COPY init.sql /docker-entrypoint-initdb.d/

# Set default encoding
ENV LANG en_US.utf8

# Use custom config file
CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]

# Expose port
EXPOSE 5432

# Add health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
    CMD pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB} || exit 1
