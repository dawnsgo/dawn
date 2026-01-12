@echo off
setlocal enabledelayedexpansion

rem 获取脚本所在目录
for %%I in ("%~dp0.") do set "directory=%%~fI"

rem 定义模块数组
set modules[0]=./
set modules[1]=./cache/redis
set modules[2]=./cache/memcache
set modules[3]=./component/http
set modules[4]=./config/consul
set modules[5]=./config/etcd
set modules[6]=./config/nacos
set modules[7]=./crypto/rsa
set modules[8]=./crypto/ecc
set modules[9]=./eventbus/kafka
set modules[10]=./eventbus/nats
set modules[11]=./eventbus/redis
set modules[12]=./locate/redis
set modules[13]=./lock/redis
set modules[14]=./lock/memcache
set modules[15]=./log/aliyun
set modules[16]=./log/tencent
set modules[17]=./network/kcp
set modules[18]=./network/tcp
set modules[19]=./network/ws
set modules[20]=./registry/consul
set modules[21]=./registry/etcd
set modules[22]=./registry/nacos
set modules[23]=./transport/rpcx
set modules[24]=./transport/grpc

rem 遍历所有模块执行go mod tidy
set count=0
:loop
if defined modules[%count%] (
    set "module=!modules[%count%]!"

    echo Processing: !module!

    rem 进入模块目录
    cd /d "!directory!\!module!"

    rem 执行go mod tidy
    go mod tidy

    if errorlevel 1 (
        echo Error processing !module!
    )

    rem 返回原始目录
    cd /d "!directory!"

    set /a count+=1
    goto loop
)

echo All modules processed successfully!
pause