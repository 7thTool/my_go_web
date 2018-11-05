FROM golang
RUN mkdir -p /go/src/beego_server
WORKDIR /go/src/beego_server
COPY . /go/src/beego_server
CMD ["go", "run"]
ONBUILD COPY . /go/src/beego_server
ONBUILD RUN go download
ONBUILD RUN go install
RUN go get github.com/astaxie/beego
RUN go get -d -v
RUN go install -v
