# urlshort

A simple HTTP server application for shortening URI addresses. It provides shortened codes for given URI addresses. Each URI address can be defined multiple times and will receive a new code each time. Data is stored in an SQLite database. The server does not record any traffic or code usage.

## Install

1. Cloning this repository to your local machine
2. OCI (Podman/Docker) image

### Build locally

This call will provide static build of source code (no external lib or binary dependencies).

> CGO_ENABLED=0 \\  
> go build -trimpath -ldflags="-s -w" -o main

The same approach is used for OCI image preparation by CI/CD pipeline.

### OCI image

Base : **scratch**

Linux systems:

> podman pull ghcr.io/vitsimon/urlshort  
> mkdir ./urlshortdata  
> podman run -p 80:8080 -v ./urlshortdata:/data ghcr.io/vitsimon/urlshort  

## Usage

Webserver is accessible from port **80** via your browser with use of your IP for external network (inside Podman network access via container alias and port 8080).  
Database with URIs and codes is stored in: **./urlshortdata/shortener.db** .

### Create short link for URI

GET  
http://your-ip:8080/c?u=https://www.seznam.cz  
> code  
> HTTP 200  

### Use short link

You will use code from **Create short link** step.
e.g.:

1. If code exists:  
GET  
http://your-ip:8080/5bd0e723ada6  
> redirection to given URI

2. If code not exists:  
GET  
http://your-ip:8080/5bd0e723ada6notexists  
> Not found  
> HTTP 404
