# php multi processor 

[![Build Status](https://travis-ci.org/sharpner/php-multiprocess-proxy.svg?branch=master)](https://travis-ci.org/sharpner/php-multiprocess-proxy)

This tool is a wrapper around the internal php webserver.
it allows multiple requests in parallel by spawning
one php -S for every request.

# Usage
```./server <port> <routerfile>```
