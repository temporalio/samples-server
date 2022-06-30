#!/bin/bash

test ! -e ca.conf && echo "ca.sh has to be run before" && exit 1

echo "Enter a unique name for the prefix of the .pem, .key, and .pfx files":
read UNIQUE_NAME
echo "Generating private key and CSR"

# Grep other conf file to grab company name.
DNS=$(grep DNS.1 ca.conf)

# Find the company name e.g. from `DNS.1=client.root.<company name>.<rand values>` 
IFS="." read -a strarr <<< "${DNS}"
CN="${strarr[3]}"

# Return a random stream of data, fold makes a new line every 4 characters, head will take the first line. 
RANDLETTER=$(cat /dev/urandom | busybox tr -dc 'a-z0-9' | busybox fold -w 4 | busybox head -n 1)

# Enter this as the DNS. 
DNS_END_ENTITY="client.endentity.${CN}.${RANDLETTER}"

GENERATED_END_ENTITY_CONF_FILE="${UNIQUE_NAME}.conf"
cat << EOF > ${GENERATED_END_ENTITY_CONF_FILE}
[req]
default_bits = 4096
default_md = sha256
req_extensions = req_ext
distinguished_name = dn
prompt = no

[req_ext]
subjectAltName = @alt_names

[dn]
O = $CN
CN = $CN client ${UNIQUE_NAME}

[alt_names]
DNS.1 = $DNS_END_ENTITY
EOF

# Generate client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout "${UNIQUE_NAME}.key" -out "${UNIQUE_NAME}-req.csr" -config "${UNIQUE_NAME}.conf"

echo "Signing certificate"

# Use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -days 365 -req -in "${UNIQUE_NAME}-req.csr" -CA "ca.pem" -CAkey "ca.key" -CAcreateserial -out "${UNIQUE_NAME}.pem" -extfile "${UNIQUE_NAME}.conf"

echo "Exporting into an unencrypted .pfx archive"
# "-keypbe NONE -certpbe NONE -passout pass:" exports into an unencrypted .pfx archive
openssl pkcs12 -export -out ${UNIQUE_NAME}.pfx -inkey ${UNIQUE_NAME}.key -in ${UNIQUE_NAME}.pem -keypbe NONE -certpbe NONE -passout pass:

# Delete the certificate signing request after the certificate has been signed. 
rm "${UNIQUE_NAME}-req.csr"

echo "Printing signed end-entity certificate"
openssl x509 -in "${UNIQUE_NAME}.pem" -noout -text

echo "---> Keep these files secure, and use ${UNIQUE_NAME}.pfx or ${UNIQUE_NAME}.pem and ${UNIQUE_NAME}.key in the SDK"
