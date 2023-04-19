FROM golang:1.19.2
COPY . /app
WORKDIR /app
RUN make setup \
&&  make generate
USER root
CMD ["go" ,"run" ,"./cmd/swagger/"]
