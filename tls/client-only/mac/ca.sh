#!/bin/bash

echo "Creating the CSR (certificate signing request)"

## This will also be used as part of the DNS for the CA and end-entity in addition to the organization setting.
echo "What is your company name?:"
read COMPANY_NAME

# Lower case company name for DNS.
CN=$(echo $COMPANY_NAME | awk '{print tolower($0)}')

## Return a random stream of data, fold makes a new line every 4 characters, head will take the first line. 
RANDLETTER=$(cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' | fold -w 4 | head -n 1)
DNS_ROOT="client.root.${CN}.${RANDLETTER}"

touch ca.conf
chmod 0755 ca.conf 

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
chmod 0755 ca.key 

openssl req -new -x509 -days 365 -key ca.key -out ca.pem -config ca.conf
chmod 0755 ca.pem 

echo "Printing root CA certificate"
openssl x509 -in ca.pem -noout -text

echo "---> Keep ca.key and ca.pem secure, but available for generating end entities in the future"