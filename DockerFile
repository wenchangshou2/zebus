

FROM golang:latest

LABEL maintainer="wcs <wenchangshou@gmail.com>"
ENV LOG_DIR=/app/logs
ENV LogLevel=info
ENV IgnoreTopic=/test,/?test*,/?web_ctrl*;
ENV MaxMsgSize=10048576
ENV MemQueueSize=100
ENV BindAddress=0.0.0.0
ENV ServerIp=192.168.20.59
ENV ServerPort=8181
ENV ETCDEnable=true
ENV ETCDCONNSTR=etcd:2379
ENV authorizationEnable=false

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build  -o zebus .
RUN echo \
[app]\\n\
LogSavePath=logs\\n/\
LogSaveName=log\\n\
LogFileExt=log\\n\
LogLevel=${LogLevel}\\n\
IgnoreTopic=${IgnoreTopic}\\n\
MaxMsgSize = 1048576\\n\
MemQueueSize = 100\\n\
[server]\\n\
BindAddress=0.0.0.0\\n\
ServerIP=${ServerIp}\\n\
ServerPort=${ServerPort}\\n\
Auth=false\\n\
AuthModel=local\\n\
AuthUsername=zoolon\\n\
AuthPassword=password\\n\
[etcd]\\n\
Enable= ${ETCDEnable}\\n\
ConnStr = ${ETCDCONNSTR}\\n\
[http]\\n\
Enable = true\\n\
Port = 9999\\n\
RunMode = debug\\n\
ReadTimeout = 3000\\n\
WriteTimeout = 3000\\n\
[authorization]\\n\
Enable = ${authorizationEnable}\\n\
>/app/conf/app.ini
VOLUME [${LOG_DIR}]
EXPOSE 8181
CMD ["./zebus"]