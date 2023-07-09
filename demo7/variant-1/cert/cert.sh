# ca.key, ca.crt
openssl genrsa \
    -out ca.key 2048
openssl req \
    -new -x509 -days 365 \
    -key ca.key \
    -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" \
    -out ca.crt

# server.key, server.csr, server.crt
openssl req \
    -newkey rsa:2048 -nodes \
    -keyout server.key \
    -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=*.example.com" \
    -out server.csr
openssl x509 -req \
    -extfile <(printf "subjectAltName=DNS:example.com,DNS:www.example.com") \
    -days 365 \
    -in server.csr \
    -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out server.crt

# ca-client.key, ca-client.crt
openssl genrsa \
    -out ca-client.key 2048
openssl req \
    -new -x509 -days 365 \
    -key ca-client.key \
    -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" \
    -out ca-client.crt

# client.key, client.csr, client.crt
openssl req \
    -newkey rsa:2048 -nodes \
    -keyout client.key \
    -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=*.client.com" \
    -out client.csr
openssl x509 -req \
    -extfile <(printf "subjectAltName=DNS:client.com,DNS:www.client.com") \
    -days 365 \
    -in client.csr \
    -CA ca-client.crt -CAkey ca-client.key -CAcreateserial \
    -out client.crt