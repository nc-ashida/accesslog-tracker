# ビルドステージ
FROM golang:1.21-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates tzdata

# 作業ディレクトリを設定
WORKDIR /app

# Goモジュールファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# 最終ステージ
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk --no-cache add ca-certificates tzdata

# 非rootユーザーを作成
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 作業ディレクトリを設定
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/main .

# 設定ファイルをコピー
COPY --from=builder /app/env.example ./env.example

# ユーザーを変更
USER appuser

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# ポートを公開
EXPOSE 8080

# アプリケーションを実行
CMD ["./main"]
