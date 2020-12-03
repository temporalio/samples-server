# This scripts generates test keys and certificates for the sample.
# In a production environment such artifacts should be genrated
# by a proper certificate authority and handled in a secure manner.

CERTS_DIR=./certs
mkdir $CERTS_DIR
CLUSTER_DIR=$CERTS_DIR/cluster
mkdir $CLUSTER_DIR
mkdir $CLUSTER_DIR/ca
mkdir $CLUSTER_DIR/accounting
mkdir $CLUSTER_DIR/development
mkdir $CLUSTER_DIR/internode
CLIENT_DIR=$CERTS_DIR/client
mkdir $CLIENT_DIR
mkdir $CLIENT_DIR/ca
mkdir $CLIENT_DIR/accounting
mkdir $CLIENT_DIR/development
TEMP_DIR=$CERTS_DIR/temp
mkdir $TEMP_DIR

generate_root_ca_cert() {
    openssl genrsa -out $2/$1.key 4096
    openssl req -new -x509 -key $2/$1.key -config $1.conf -days 365 -out $2/$1.pem
}

generate_cert() {
    openssl genrsa -out $2/$1.key 4096
    openssl req -new -key $2/$1.key -out $TEMP_DIR/$1.csr -config $1.conf
    openssl x509 -req -in $TEMP_DIR/$1.csr -CA $3.pem -CAkey $3.key -CAcreateserial -out $2/$1.pem -days 365 -extfile $1.conf -extensions $4
    if [[ $5 != 'no_chain' ]]
    then
      cat $2/$1.pem $3.pem > $2/$1-chain.pem
    fi
}

echo Generate a private key and a certificate for server root CA
generate_root_ca_cert server-root-ca $CLUSTER_DIR/ca

echo Generate a private key and a certificate for server intermediate CA
generate_cert server-intermediate-ca $CLUSTER_DIR/ca $CLUSTER_DIR/ca/server-root-ca v3_ca no_chain

echo Generate a private key and a certificate for internode communication 
generate_cert cluster-internode $CLUSTER_DIR/internode $CLUSTER_DIR/ca/server-intermediate-ca req_ext

echo Generate a private key and a server certificate for accounting namespace
generate_cert cluster-accounting $CLUSTER_DIR/accounting $CLUSTER_DIR/ca/server-intermediate-ca req_ext

echo Generate a private key and a server certificate for development namespace
generate_cert cluster-development $CLUSTER_DIR/development $CLUSTER_DIR/ca/server-intermediate-ca req_ext


echo Generate a private key and a certificate for client root CA
generate_root_ca_cert client-root-ca $CLIENT_DIR/ca

echo Generate a private key and a certificate for client intermediate CA for accounting namespace
generate_cert client-intermediate-ca-accounting $CLIENT_DIR/ca $CLIENT_DIR/ca/client-root-ca v3_ca no_chain

echo Generate a private key and a certificate for client intermediate CA for development namespace
generate_cert client-intermediate-ca-development $CLIENT_DIR/ca $CLIENT_DIR/ca/client-root-ca v3_ca no_chain

echo Generate a private key and a certificate for accounting namespace client 
generate_cert client-accounting-namespace $CLIENT_DIR/accounting $CLIENT_DIR/ca/client-intermediate-ca-accounting req_ext

echo Generate a private key and a certificate for development namespace client 
generate_cert client-development-namespace $CLIENT_DIR/development $CLIENT_DIR/ca/client-intermediate-ca-development req_ext
