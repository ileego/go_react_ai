#!/bin/bash
# 等待数据库就绪后再执行后续命令
# 用法: ./wait-for-db.sh [host] [port] [timeout_seconds]

HOST=${1:-localhost}
PORT=${2:-5432}
TIMEOUT=${3:-30}

echo "等待 PostgreSQL ($HOST:$PORT) 就绪..."

for i in $(seq 1 $TIMEOUT); do
    if pg_isready -h "$HOST" -p "$PORT" > /dev/null 2>&1; then
        echo "数据库已就绪"
        exit 0
    fi
    echo "第 $i 秒，数据库未就绪，继续等待..."
    sleep 1
done

echo "超时：数据库在 ${TIMEOUT} 秒内未就绪"
exit 1
