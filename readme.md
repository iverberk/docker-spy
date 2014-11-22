# Introduction

Docker-spy provides a DNS service based on Docker container events. It keeps an in-memory database of records that map container hostnames to ip addresses. When containers are start/stopped/destroyed it keeps track of their location.

It is specifically targeted at small local development environments where you want an easy way to connect with your containers.

# Usage

The easiest way to run docker-spy is through Docker. The [image](https://registry.hub.docker.com/u/iverberk/docker-spy/) is based on the scratch image (basically a zero sized image) and contains only the compiled Go executable. 

### configuration

Docker-spy can be configured through a number of environment variables:

* DNS_BIND: the address that the DNS server will bind to in the container. Defaults to '0.0.0.0'.
* DNS_PORT: the port on which the DNS server will be reachable (tcp/udp). Defaults to '53'
* DNS_RECURSOR: the recursor to use when DNS requests are made to non-local addresses. Defaults to '8.8.8.8' from the Dockerfile
* DNS_DOMAIN: the domain that docker-spy should consider local and keep records for. Defaults to 'localdomain'
* DOCKER_HOST: the location of the Docker daemon. Defaults to the DOCKER_HOST environment variable or, if the DOCKER_HOST environment variable is not set, unix:///var/run/docker.sock. Setting this explicitly allows you to override the location.
Docker-spy can be started with the following command (using all defaults, add any environment settings that you wish to change):

### DNS Forwarding

Docker-spy will consider all DNS requests that end with the above configured DNS_DOMAIN to be internal requests that should be mapped to a container. All other DNS requests are forwarded to the recursor, so the DNS server should be relatively transparent.

### Prerequisites

Before starting docker-spy you should know the following things about your system:

1. The IP address of the Docker bridge. Issue an ```ifconfig``` and look for the Docker0 bridge entry. OSX users should first ssh into the boot2docker virtual machine with ```boot2docker ssh```
2. (OSX Only) The IP address of the boot2docker virtual machine. Run ```boot2docker ip``` to find out what it is.

### Running Docker-spy
To run docker-spy you can issue the following command:

```
docker run --name docker_spy -p 53:53/udp -p 53:53 -v /var/run/docker.sock:/var/run/docker.sock iverberk/docker-spy
```

This maps the Docker socket as a volume in the container so that events may be tracked and it publishes port 53 on udp/tcp to the host. Add a ```-d``` parameter to run the container in the background. You may then inspect the logs with ```docker logs docker_spy```

### OSX (boot2docker)

To have seamless DNS resolution and access to your containers you should perform the following steps:

1. Create an /etc/resolver/$DOMAIN file (create the /etc/resolver directory first). $DOMAIN should be substituted with the domain that you use for local development (e.g. for 'localdomain' create a /etc/resolver/localdomain file). Add the following contents to this file: ```nameserver x.x.x.x``` (x.x.x.x should be replaced with the Docker bridge IP address that you looked up earlier)
2. Create a route so that the container ip range is accessible from the osx host system through the boot2docker host adapter: ```sudo route -n add -net 172.17.0.0 192.168.59.104```<br>**important:** substitute 192.168.59.104 with the ip address of **your** boot2docker vm (run boot2docker ip to find out what it is). Substitute 172.17.0.0 with the appropriate range (if you have the Docker bridge IP just chop of the last two digits and replace them with zeros, e.g. if the Docker bridge IP is 10.0.42.1 then you should use 10.0.0.0 as **your** -net parameter in the above command)
3. You should now be able to ping your containers directly and run DNS queries against the Docker bridge IP. **important**: check for any firewall rules on the host that may be blocking the udp/tcp traffic on port 53!

### Linux

To have automatic DNS resolution of your containers you should update your /etc/resolv.conf and the Docker bridge IP address as a resolver (usually 172.17.42.1 but check it with ifconfig to be certain).

# Issues and Contributing

Docker-spy is really young and has a lot of rough edges. I really wanted to have a basic, working solution before adding nice-to-haves. It is also my first Go program so there will probably be less then idiomatic constructs in the program. Fixes and enhancements are gladly excepted. Issues may be filed against the github repository.

### Building

To build docker-spy just install the go build environment and run ```go build -o docker-spy *.go``` in the directory.