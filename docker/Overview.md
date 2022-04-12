
# What is the go-dots

"go-dots" is a DDoS Open Threat Signaling (dots) implementation written in Go. This implmentation is based on the Internet drafts below. 

* RFC 8782 (was draft-ietf-dots-signal-channel)
* RFC 8783 (was draft-ietf-dots-data-channel)
* draft-ietf-dots-architecture-18
* RFC 8612 (was draft-ietf-dots-requirements)
* draft-ietf-dots-use-cases-21
* draft-ietf-dots-signal-filter-control-04
* draft-ietf-dots-signal-call-home-07

This implementation is not fully compliant with the documents listed above.  For example, we are utilizing CoAP as the data channel protocol while the current version of the data channel document specifies RESTCONF as the data channel protocol.

Licensed under Apache License 2.0.


# How to use this image

## Install kubernetes

To install kubernetes, refer to the following link:
* [kubernetes-ubuntu](https://matthewpalmer.net/kubernetes-app-developer/articles/install-kubernetes-ubuntu-tutorial.html)

## Structure the configuration files

- The dots folder with the structure as:
    ```
    ~/dots
        |--/certs
                |-- /certs/* files
        |--/config
                |-- /test_dump.sql
                |-- /gobgpd.conf
                |-- /dots_server.yaml
                |-- /dots_client.yaml
    ```

    ```
    # Create the certificate folder
    mkdir -p ~/dots/certs

    # Create the configuration folder
    mkdir -p ~/dots/config
    ```

- Mounting the dots folder from host directory to minikube. Then the host directory and the virtual machine minikube contain the dots folder with the files and the values of files that are same.
    ```
    $ minikube mount ~/dots:/dots
    ```

- Created the go-dots server, the mysql and the gobgp on Kubernetes by the [DeploymentServer.yaml](https://github.com/nttdots/go-dots/blob/master/docker/DeploymentServer.yaml) file. 

    ```
    $ curl -OL https://raw.githubusercontent.com/nttdots/go-dots/master/docker/DeploymentServer.yaml
    ```
- Created the go-dots client on Kubernetes by the [DeploymentClient.yaml](https://github.com/nttdots/go-dots/blob/master/docker/DeploymentClient.yaml) file. 

    ```
    $ curl -OL https://raw.githubusercontent.com/nttdots/go-dots/master/docker/DeploymentClient.yaml
    ```

# Using Docker container

To use the go-dots server, the go-dots client, the mysql and go-bgp. Following as below:

- Created Deployment server (Ran the go-dots server and gobgp)
    ```
    $ kubectl create -f DeploymentServer.yaml
    ```

- Import data into mysql container
    ```
    $ cd ~/dots/config
    $ kubectl exec -it deployment/server -c mysql -- mysql -uroot -proot dots < test_dump.sql
    ```

- Get Pod to check ip address of the go-dots server
    ```
    $ kubectl get pods --output=wide
    ```

- Setting environment parameters in configmap
    ```
    $ kubectl create configmap client-config --from-literal SERVER_IP_ADDRESS=172.17.0.7 --from-literal SERVER_PORT=4646
    ```

- Created Deployment client (Ran the go-dots client)
    ```
    $ kubectl create -f DeploymentClient.yaml
    ```

- Executing the go-dots client controller
    ```
    $ kubectl exec -it deployment/go-dots-client /bin/bash
    ```

- To get logs of go-dots server/client. Using kubectl command
    ```
    $ kubectl logs deployment/server -c go-dots-server
    $ kubectl logs deployment/go-dots-client
    ```

- After modify go-dots config, you restart deployment
    ```
    $ kubectl rollout restart deployment/server
    ```
    or
    ```
    $ kubectl rollout restart deployment/go-dots-client
    ```

## GoBGP

Check the route is installed successfully in gobgp server

    $ gobgp global rib

    ```
    Network              Next Hop             AS_PATH              Age        Attrs
    *> 172.16.238.100/32    172.16.238.254                            00:00:42   [{Origin: i}]
    ```

Check the flowspec route is installed successfully in gobgp server

    $ gobgp global rib -a ipv4-flowspec
    $ gobgp global rib -a ipv6-flowspec

    ```
    Network                                                                   Next Hop      AS_PATH    Age        Attrs
    *> [destination: 1.1.2.0/24][protocol: ==tcp][destination-port: >=443&<=800] fictitious               00:00:06   [{Origin: i} {Extcomms: [redirect: 1.1.1.0:100]}]
    ```

##  Signal Channel
The primary purpose of the signal channel is for a DOTS client to ask a DOTS server for help in mitigating an attack, and for the DOTS server to inform the DOTS client about the status of such mitigation.

### Client Controller [mitigation_request]

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Put \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123 \
     -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleMitigationRequestDraft.json

In order to handle out-of-order delivery of mitigation requests, 'mid' values MUST increase monotonically. Besides, if the 'mid' value has exceeded 3/4 of (2**32 - 1), it should be reset by sending a mitigation request with 'mid' is set to '0' to avoid 'mid' rollover. However, the reset request is only accepted by DOTS server at peace-time (have no any active mitigation request which is maintaining).

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Put \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=0 \
     -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleMitigationRequestDraft.json

### Client Controller [mitigation_retrieve_all]

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Get \
     -cuid=dz6pHjaADkaFTbjr0JGBpw

### Client Controller [mitigation_retrieve_one]

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Get \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123

### Client Controller [mitigation_withdraw]

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Delete \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123

### Client Controller [mitigation_observe]
A DOTS client can convey the 'observe' option set to '0' in the GET request to receive notification whenever status of mitigation request changed and unsubscribe itself by issuing GET request with 'observe' option set to '1'

Subscribe for resource observation:

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Get \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123 -observe=0

Unsubscribe from resource observation:

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Get \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123 -observe=1

Subscriptions are valid as long as current session exists. When session is renewed (e.g DOTS client does not receive response from DOTS server for its Ping message in a period of time, it decided that server has been disconnected, then re-connects), all previous subscriptions will be lost. In such cases, DOTS clients will have to re-subscribe for observation. Below is recommended step: 

    ・GET a list of all existing mitigations (that were created before server restarted)
    ・PUT mitigations  one by one
    ・GET + Observe for all the mitigations that should be observed

### Client Controller [mitigation_efficacy_update]
A DOTS client can convey the 'If-Match' option with empty value in the PUT request to transmit DOTS mitigation efficacy update to the DOTS server:

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Put \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123 -ifMatch="" \
     -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleMitigationRequestDraftEfficacyUpdate.json

### Client Controller [session_configuration_request]

    $ $GOPATH/bin/dots_client_controller -request session_configuration -method Put \
     -sid 234 \
     -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleSessionConfigurationDraft.json

In order to handle out-of-order delivery of session configuration, 'sid' values MUST increase monotonically.

### Client Controller [session_configuration_retrieve_default]

    $ $GOPATH/bin/dots_client_controller -request session_configuration -method Get

### Client Controller [session_configuration_retrieve_one]

    $ $GOPATH/bin/dots_client_controller -request session_configuration -method Get \
      -sid 234

### Client Controller [session_configuration_delete]

    $ $GOPATH/bin/dots_client_controller -request session_configuration -method Delete \
      -sid 234

###  Client Controller [client_configuration_request]
DOTS signal channel session configuration supports 2 sets of parameters : 'mitigating-config' and 'idle-config'.
The same or distinct configuration set may be used during times when a mitigation is active ('mitigating-config') and when no mitigation is active ('idle-config').
Dots_client uses 'idle-config' parameter set by default. It can be configured to switch to the other parameter set by client_configuration request

Configure dots_client to use 'idle-config' parameters

    $ $GOPATH/bin/dots_client_controller -request client_configuration -method POST \
    -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleClientConfigurationRequest_Idle.json

Configure dots_client to use 'mitigating-config' parameters

    $ $GOPATH/bin/dots_client_controller -request client_configuration -method POST \
    -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleClientConfigurationRequest_Mitigating.json

##  Data Channel
The primary purpose of the data channel is to support DOTS related configuration and policy information exchange between the DOTS client and the DOTS server.

All shell-script and sample json files are located in below directory:
    $ cd $GOPATH/src/github.com/nttdots/go-dots/dots_client/data/

### Get Root Resource Path

    Get root resource:
    $ ./get_root_resource.sh SERVER_NAME

    Example:
        - Request: $ ./get_root_resource.sh https://127.0.0.1:10443
        - Response:
        <XRD xmlns="https://127.0.0.1">
            <Link rel="restconf" href="https://127.0.0.1:10443/v1/restconf"></Link>
        </XRD>

    "{href}" value will be used as the initial part of the path in the request URI of subsequent requests

### Managing DOTS Clients
Registering DOTS Clients

    Post dots_client:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} POST {href}/data/ietf-dots-data-channel:dots-data sampleClient.json

    Put dots_client:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} PUT {href}/data/ietf-dots-data-channel:dots-data/dots-client=123 sampleClient.json

Uregistering DOTS Clients

    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} DELETE {href}/data/ietf-dots-data-channel:dots-data/dots-client=123

### Managing DOTS Aliases
Create Aliases

    Post alias:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} POST {href}/data/ietf-dots-data-channel:dots-data/dots-client=123 sampleAlias.json

    Put alias:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} PUT {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases/alias=xxx sampleAlias.json

Retrieve Installed Aliases

    Get all aliases without 'content' parameter (default is get all type attributes, including configurable and non-configurable attributes):
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases

    Get all aliases with 'content'='config' (get configurable attributes only):
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases?content=config

    Get all aliases with 'content'='nonconfig' (get non-configurable attributes only):
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases?content=nonconfig

    Get all aliases with 'content'='all'(get all type attributes, including configurable and non-configurable attributes):
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases?content=all

    Get specific alias without 'content' parameter:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases/alias=https1

    Get specific alias with 'content'='config':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases/alias=https1?content=config

    Get specific alias with 'content'='nonconfig':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases/alias=https1?content=nonconfig

    Get specific alias with 'content'='all':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases/alias=https1?content=all

Delete Aliases

    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} DELETE {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/aliases/alias=https1

### Managing DOTS Filtering Rules
Retrieve DOTS Filtering Capabilities

    Get Capabilities without 'content' parameter:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/capabilities

    Get Capabilities with 'content'='config':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/capabilities?content=config

    Get Capabilities with 'content'='nonconfig':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/capabilities?content=nonconfig

    Get Capabilities with 'content'='all':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/capabilities?content=all

Install Filtering Rules

    Post acl:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} POST {href}/data/ietf-dots-data-channel:dots-data/dots-client=123 sampleAcl.json

    Put acl:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} PUT {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls/acl=xxx sampleAcl.json

Retrieve Installed Filtering Rules

    Get all Acl without 'content' parameter:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls

    Get all Acl with 'content'='config':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls?content=config

    Get all Acl with 'content'='nonconfig':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls?content=nonconfig

    Get all Acl with 'content'='all':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls?content=all

    Get specific acl without 'content' parameter:
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls/acl=sample-ipv4-acl

    Get specific acl with 'content'='config':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls/acl=sample-ipv4-acl?content=config

    Get specific acl with 'content'='nonconfig':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls/acl=sample-ipv4-acl?content=nonconfig

    Get specific acl with 'content'='all':
    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} GET {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls/acl=sample-ipv4-acl?content=all

Remove Filtering Rules

    $ ./do_request_from_file.sh --client-cert {client-cert-path} --client-key {client-key-path} --ca-cert {ca-cert-path} DELETE {href}/data/ietf-dots-data-channel:dots-data/dots-client=123/acls/acl=sample-ipv4-acl

## Signal Channel Control Filtering
Unlike the DOTS signal channel, the DOTS data channel is not expected to deal with attack conditions.
Therefore, when DOTS client is under attacked by DDoS, the DOTS client can use DOTS signal channel protocol to manage the filtering rule in DOTS Data Channel to enhance the protection capability of DOTS protocols.

### Client Controller [mitigation_control_filtering]

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Put \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123 \
     -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleMitigationRequestDraftControlFiltering.json

## Signal Channel Call Home
The DOTS signal channel Call Home identify the source to block DDoS attack traffic closer to the source(s) of a DDoS attack.
when the DOTS client is under attacked by DDoS, the DOTS client sends the attack traffic information to the DOTS server. The DOTS server in turn uses the attack traffic information to identify the compromised devices launching the outgoing DDoS attack and takes appropriate mitigation action.

### Client Controller [mitigation_call_home]

    $ $GOPATH/bin/dots_client_controller -request mitigation_request -method Put \
     -cuid=dz6pHjaADkaFTbjr0JGBpw -mid=123 \
     -json $GOPATH/src/github.com/nttdots/go-dots/dots_client/sampleMitigationRequestDraftCallHome.json


# Certificate Configuration

### Precondition
* The GnuTLS has been installed.

### Configure The Certificate

Typically, the client's/server's certificate is a single identifier type which means that the certificate has only one common name (CN-ID) as identifier. However, the Common Name is not strongly typed because the Common Name can contain a human-friendly string, not a DNS domain name. Moreover, the client's/server's certificate can be multi identifiers type which including the Common Name (CN-ID) and some Subject Alternative Name (DNS-ID, SRV-ID), in order to ensure that at least one DNS qualified domain name. In go-dots, two certificate types are configured as below:

* The single identifier type

   The template file only has the Common Name (CN-ID)

   Example: In file [template_client](./certs/template_client.txt), configure as follow:
    ```
    # X.509 server certificate options

    organization = "sample client"
    state = "Tokyo"
    country = JP
    cn = "client.sample.example.com"
    expiration_days = 365

    # X.509 v3 extensions
    signing_key
    encryption_key
    ```

* The multi identifiers type

    The template file has the Common Name (CN-ID) and some Subject Alternative Name (DNS-ID, SRV-ID)

    Example: In file [template_client](./certs/template_client.txt), add more Subject Alternative Name as follow:

    ```
    # DNS name(s) of the server
    dns_name = "xample1.example.com"
    dns_name = "_xampp.example.com
    ```

To add/change the server's certificate or the client's certificate, execute with following command:
```
./update_keys.sh
```

For more detailed information about configuration of the certificate and the GnuTLS, refer to the following link:
* [action_gnutls_scripted.md](https://gist.github.com/epcim/832cec2482a255e3f392)
