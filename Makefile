frontend:
	$(MAKE) -C frontend all

proxy:
	$(MAKE) -C proxy all

.DEFAULT_GOAL := all
.PHONY: all frontend proxy
all: frontend proxy

dockerfiles:
	@$(MAKE) -f Makefile.docker dockerfiles
	@sh -c "ls -1 */Dockerfile.*"

install:
	install -m 755 frontend/frontend /usr/local/bin/frontend
	install -m 755 proxy/proxy /usr/local/bin/proxy
