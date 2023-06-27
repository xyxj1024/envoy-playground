openssl genrsa -out private.key 4096
openssl req -new -x509 -key private.key -out certificate.pem -days 365