# Use the official PostgreSQL image as the base image
FROM postgres:latest

# Set environment variables
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres123
ENV POSTGRES_DB=sothea

# Copy the initialization scripts into the Docker entrypoint directory
COPY ./sql /docker-entrypoint-initdb.d

# Expose the PostgreSQL port
EXPOSE 5432

# Set the entrypoint to the default PostgreSQL entrypoint
ENTRYPOINT ["docker-entrypoint.sh"]

# Run the PostgreSQL server
CMD ["postgres"]
