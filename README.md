# Couch-SSO

## Introduction 

```
                          +----------------+
                          |                |
                          |  Token Server  |
                          |                |
                          +----------------+ 
                              ^        |
                              |        | ( Access Token )
   ( Retrieve/Refresh Token ) |        | 
                              |        | 
                              |        | 
+-----------------------------|--------|--------------+                         +------------------------------------------------------+
|               ( Client )    |        |              |                         |                            ( Cloud )                 |
|                             |        v              |        ( JWT )          |                                                      |
| +-----------+            +----------------------+   |------------------------>|  +----------------------+              +-----------+ |
| |           |(Basic Auth)|                      |   |                         |  |    Couch-SSO Server  | (Basic Auth) |           | |
| |  CouchDB  +------------+    Couch-SSO Client  +---+    ( Public network )   +--+                      +<------------>+  CouchDB  | |
| |           |            |                      |   |                         |  |   ( Authorization )  |              |           | |
| +-----------+            +----------------------+   |<------------------------|  +----------------------+              +-----------+ |
|                                                     |                         |                                                      |
+-----------------------------------------------------+                         +------------------------------------------------------+

```
An application implementation of [JSON Web Token](https://tools.ietf.org/html/rfc7519).   
[Couch-SSO](http://github.com/dravenk/couch-sso) can be deployed on both the server side and the client side.You can also write your own client and use couch-sso on the server side alone.   
The client change the Authorization in the request Header to JWT and forwards it to the Authorization Server.  
After the Authorization Server listens to the request, check whether the request information in the Header contains JWT.   
Verify the validity of token by public.Pem, and choose whether to allow access to the resources of the server-side [CouchDB](http://couchdb.apache.org/).   

## Document description 
```
example.public.pem               // The public key used to verify the token signature is obtained from the token server.
example-client.config.json
example-server.config.json
```

## Deploy

### Use with docker in server side. 
1. Fill out your configuration.  
    `cp example-server.config.json config.json && cp example-.public.pem public.pem`
2. Build docker image by default docker image.    
    `docker build -t couch-sso .`    
3. Start the container.   
    `docker run --name couch-sso -p 8008:8008 -d couch-sso:latest`  

**Note.** 
You can also deploy on the client side. Just use `client` instead of `server`.

Example. 
--
### Use with CouchDB in client.  

The username and password in the client configuration file config.json are not only used to retrieve the token from token server, but also the authentication account and password of couchdb-sso.
Supports grant_type `password` and `client_credentials`. 

1. copy the client-client.config.json to config.json  
    `cp client-client.config.json config.json`  
    
    ```
    {
      "proxy_port": "8008",
      "token_host": "https://example.com/oauth/token",
      "username": "server_user",
      "password": "server_pass",
      "client_id": "client_id-rand",
      "client_secret": "client_secret-rand",
      "grant_type": "password",
      "remote": "https://sso.exmpale.com:5984"
    }
    ```
2. build docker image by default docker image.    
    `docker build -t couch-sso .`    
3. run docker image.   
    `docker run --name couch-sso -p 8008:8008 -d couch-sso:latest`  
4. Check the client host's IP address.  
    `ifconig`  
    example.The IP address is `192.168.1.120`.  
5. New Replication      
    Create a New Replication.Open browser in  `http://192.168.1.120:5984/_utils/#/replication`  
    Since the custom CouchDB database name is db_name, the address is entered as localhost IP.  
    `http://server_use:server_pass@192.168.1.120:8008/db_name`  
