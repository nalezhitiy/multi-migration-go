# ==========================================================
# GLOBALS ARGUMENTS
# ==========================================================
ARG ALPINE_VER=3.24
ARG GOLANG_VER=1.25.12

ARG BUILD_ARTIFACT=app
ARG BUILD_WORKDIR=/build
ARG WORKDIR=/app


# ==========================================================
# BASE STAGE
# ==========================================================
FROM golang:${GOLANG_VER}-alpine${ALPINE_VER} AS base
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV TZ=UTC

RUN apk update --no-cache && apk add --no-cache tzdata git


# ==========================================================
# DEVELOPMENT STAGE
# ==========================================================
FROM base AS dev

ARG WORKDIR

RUN apk add --no-cache build-base

WORKDIR $WORKDIR

COPY ./src/go.mod ./ ./src/go.sum* ./

RUN go mod download


# ==========================================================
# BUILDER STAGE
# ==========================================================
FROM base AS backend-builder

ARG BUILD_ARTIFACT
ARG BUILD_WORKDIR

WORKDIR ${BUILD_WORKDIR}

COPY ./src/go.mod ./ ./src/go.sum* ./
RUN go mod download

COPY ./ ./
RUN go build -o ${BUILD_ARTIFACT} ./app.go


# ==========================================================
# PRODUCTION STAGE
# ==========================================================
ARG ALPINE_VER
FROM alpine:${ALPINE_VER} AS prod

ARG BUILD_ARTIFACT
ARG BUILD_WORKDIR
ARG WORKDIR

ARG USER_NAME=appuser
ARG USER_UID=1000

ENV TZ=UTC
ENV APP_ENTRYPOINT=${WORKDIR}/${BUILD_ARTIFACT}

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata && \
    addgroup -g ${USER_UID} ${USER_NAME} && \
    adduser -D -u ${USER_UID} -G ${USER_NAME} ${USER_NAME}

WORKDIR ${WORKDIR}

COPY --from=backend-builder --chown=${USER_NAME}:${USER_NAME} ${BUILD_WORKDIR}/${BUILD_ARTIFACT} ${APP_ENTRYPOINT}

USER ${USER_NAME}

ENTRYPOINT ["/bin/sh", "-c", "${APP_ENTRYPOINT} \"$@\"", "--"]
