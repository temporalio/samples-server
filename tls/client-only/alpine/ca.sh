#!/bin/bash

echo "Creating the CSR (certificate signing request)"

## This will also be used as the DNS for the end-entity
echo "What is your company name?:"
read COMPANY_NAME

CN="${COMPANY_NAME,,}"

## Return a random stream of data, fold makes a new line every 4 characters, head will take the first line. 
RANDLETTER=$(cat /dev/urandom | busybox tr -dc 'a-z0-9' | busybox fold -w 4 | busybox head -n 1)

DNS_ROOT="client.root.${CN}.${RANDLETTER}"

GENERATED_ROOT_CONF_FILE="ca.conf"
cat << EOF > ${GENERATED_ROOT_CONF_FILE}
[req]
default_bits = 4096
default_md = sha256
req_extensions = req_ext
x509_extensions = v3_ca
distinguished_name = dn
prompt = no

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical,CA:TRUE,pathlen:0
keyUsage = critical,digitalSignature,cRLSign,keyCertSign

[req_ext]
subjectAltName = @alt_names

[dn]
O = $COMPANY_NAME

[alt_names]
DNS.1 = $DNS_ROOT

EOF

openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.pem -config ca.conf

echo "Printing root CA certificate"
openssl x509 -in ca.pem -noout -text

echo "---> Keep ca.key and ca.pem secure, but available for generating end entities in the future"