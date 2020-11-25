# TEMPORAL_TLS_CERTS_DIR is used in docker-compose.yml to point 
# to the location of generated test certificates within the container
export TEMPORAL_TLS_CERTS_DIR=/etc/temporal/config/certs

# TEMPORAL_LOCAL_CERT_DIR is used in docker-compose.yml to point
# to our local directory with generated certificates to be mounted
# as a volume at TEMPORAL_TLS_CERTS_DIR within the container
export TEMPORAL_LOCAL_CERT_DIR=./certs

docker-compose up