# Global ARGs
ARG FRONTEND_DIR=webssr
ARG PNPM_VERSION=10.18.1

# ----------------------------------------------------
# Stage 1: Frontend Build
# ----------------------------------------------------
FROM node:22-alpine AS frontend
ARG FRONTEND_DIR
ARG PNPM_VERSION

ENV PNPM_HOME=/root/.local/share/pnpm
ENV PNPM_STORE_DIR=${PNPM_HOME}/store
ENV PATH="${PNPM_HOME}:${PATH}"

RUN corepack enable && corepack prepare "pnpm@${PNPM_VERSION}" --activate

WORKDIR /app/${FRONTEND_DIR}

COPY ${FRONTEND_DIR}/package.json ./package.json
COPY ${FRONTEND_DIR}/pnpm-lock.yaml ./pnpm-lock.yaml

RUN --mount=type=bind,source=${FRONTEND_DIR},target=/tmp/src,ro \
    if [ -f /tmp/src/pnpm-workspace.yaml ]; then cp /tmp/src/pnpm-workspace.yaml ./pnpm-workspace.yaml; fi

RUN --mount=type=cache,target=${PNPM_STORE_DIR} \
    pnpm install --frozen-lockfile

COPY ${FRONTEND_DIR} .

RUN pnpm build


# ----------------------------------------------------
# Stage 2: Backend Build (使用 Debian-based 镜像进行 CGO 编译)
# ----------------------------------------------------
FROM golang:1.25 AS backend

ARG TARGETOS
ARG TARGETARCH

# 安装 C/C++ 编译工具链，用于 CGO
RUN apt-get update && apt-get install -y \
    gcc \
    g++

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

COPY ./admin ./admin
COPY ./api ./api
COPY ./conf ./conf
COPY ./dao ./dao
COPY ./job ./job
COPY ./pkg ./pkg
COPY main.go .

# FIX: 确保从 frontend 阶段复制构建产物的路径正确。
# frontend 阶段的 WORKDIR 是 /app/webssr，所以 dist 目录在其内部。
COPY --from=frontend /app/${FRONTEND_DIR}/dist ./dist 

ENV CGO_ENABLED=1
RUN GOOS=${TARGETOS:-$(go env GOOS)} \
    GOARCH=${TARGETARCH:-$(go env GOARCH)} \
    go build -o build/server -ldflags "-w -s" .


# ----------------------------------------------------
# Stage 3: Final Runtime (FIX: 切换到 Alpine 并安装 GLIBC 兼容包)
# ----------------------------------------------------
FROM alpine:latest AS final

WORKDIR /app

# FIX: 安装 GLIBC 兼容包，这是解决 libm.so.6: version `GLIBC_2.38' not found 错误的关键。
# libc6-compat: 提供 GLIBC 兼容层。
# gcompat: 提供 GNU C 库的兼容性。
# libstdc++: 解决 libstdc++.so.6 的依赖。
RUN apk update && \
    apk add --no-cache libc6-compat gcompat libstdc++ tzdata && \
    rm -rf /var/cache/apk/*

# 复制 Go 可执行文件
COPY --from=backend /app/build/server /app/

# 复制前端静态资源
COPY --from=backend /app/dist /app/dist

# 设置时区（可选）
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

EXPOSE 8080

CMD ["/app/server"]