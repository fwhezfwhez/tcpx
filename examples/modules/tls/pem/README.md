```json
openssl genrsa > key.pem
openssl req -new -x509 -key key.pem > cert.pem


You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:CN
State or Province Name (full name) [Some-State]:JX
Locality Name (eg, city) []:nc
Organization Name (eg, company) [Internet Widgits Pty Ltd]:tcpx
Organizational Unit Name (eg, section) []:tcpx
Common Name (e.g. server FQDN or YOUR name) []:localhost
Email Address []:null

```