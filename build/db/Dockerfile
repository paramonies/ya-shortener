FROM postgres:13.2-alpine as build

RUN apk update && \
apk add postgresql-dev libpq git build-base llvm clang && \
mkdir -p /usr/src && \
cd /usr/src && \
git clone https://github.com/theory/pg-semver && \
cd pg-semver && \
make && \
make install