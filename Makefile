frontend:
	$(MAKE) -C frontend all

proxy:
	$(MAKE) -C proxy all

.DEFAULT_GOAL := all
.PHONY: all frontend proxy
all: frontend proxy

install:
	install -m 755 frontend/frontend /usr/local/bin/bird-lg-go
	install -m 755 proxy/proxy /usr/local/bin/bird-lgproxy-go
