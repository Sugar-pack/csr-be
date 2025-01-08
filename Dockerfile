FROM golang:1.22.5
COPY . /app
WORKDIR /app
RUN make setup \
&&  make generate
USER root
CMD ["go" ,"run" ,"./cmd/swagger/"]
